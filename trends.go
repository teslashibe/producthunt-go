package producthunt

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
)

// AnalyzeTrends paginates through posts in the given date range and computes
// aggregate analytics. Default range: last 7 days.
//
// The function respects ctx cancellation and returns a partial TrendReport
// alongside ErrPartialResult if cancelled mid-scrape.
func (c *Client) AnalyzeTrends(ctx context.Context, opts ...TrendOption) (*TrendReport, error) {
	o := applyTrendOptions(opts)

	stopSet := buildStopSet(o.stopWords)

	var allPosts []Post
	cursor := ""
	partial := false

	for {
		if ctx.Err() != nil {
			partial = true
			break
		}

		listOpts := []PostListOption{
			WithPostedAfter(o.from),
			WithPostedBefore(o.to),
			WithPostsOrder(PostsOrderNewest),
			WithPostListFirst(50),
		}
		if cursor != "" {
			listOpts = append(listOpts, WithPostListAfter(cursor))
		}
		if len(o.topics) > 0 {
			listOpts = append(listOpts, WithPostsTopic(o.topics[0]))
		}

		page, err := c.GetPostsByDate(ctx, o.from, listOpts...)
		if err != nil {
			if ctx.Err() != nil {
				partial = true
				break
			}
			return nil, err
		}

		allPosts = append(allPosts, page.Items...)
		if !page.HasNext || page.EndCursor == "" {
			break
		}
		cursor = page.EndCursor
	}

	report := computeTrendReport(allPosts, o, stopSet)

	if partial {
		return report, fmt.Errorf("%w", ErrPartialResult)
	}
	return report, nil
}

func computeTrendReport(posts []Post, o trendOptions, stopSet map[string]bool) *TrendReport {
	report := &TrendReport{
		DateRange:     [2]time.Time{o.from, o.to},
		PostsAnalyzed: len(posts),
	}

	if len(posts) == 0 {
		return report
	}

	keywords := map[string]int{}
	topicCounts := map[string]*TopicFreq{}
	dayOfWeek := map[time.Weekday]int{}
	makerMap := map[string]*makerAccum{}

	for _, p := range posts {
		report.TotalVotes += p.VotesCount
		report.TotalComments += p.CommentsCount

		extractKeywords(p.Name+" "+p.Tagline, keywords, stopSet)

		for _, t := range p.Topics {
			if _, ok := topicCounts[t.Slug]; !ok {
				topicCounts[t.Slug] = &TopicFreq{Slug: t.Slug, Name: t.Name}
			}
			topicCounts[t.Slug].Count++
		}

		dayOfWeek[p.CreatedAt.Weekday()]++

		for _, m := range p.Makers {
			if _, ok := makerMap[m.ID]; !ok {
				makerMap[m.ID] = &makerAccum{
					UserID:   m.ID,
					Username: m.Username,
					Name:     m.Name,
				}
			}
			ma := makerMap[m.ID]
			ma.Products++
			ma.TotalVotes += p.VotesCount
		}
	}

	report.AvgVotes = float64(report.TotalVotes) / float64(len(posts))
	report.AvgComments = float64(report.TotalComments) / float64(len(posts))

	// Top products by votes
	sorted := make([]Post, len(posts))
	copy(sorted, posts)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].VotesCount > sorted[j].VotesCount })
	topN := o.topN
	if topN > len(sorted) {
		topN = len(sorted)
	}
	for _, p := range sorted[:topN] {
		report.TopProducts = append(report.TopProducts, PostSummary{
			ID: p.ID, Name: p.Name, Tagline: p.Tagline,
			Votes: p.VotesCount, Comments: p.CommentsCount,
			DailyRank: p.DailyRank, URL: p.URL, Website: p.Website,
		})
	}

	// Rising products: high votes relative to recency
	report.RisingProducts = findRisingProducts(posts, o.topN)

	// Top keywords
	report.TopKeywords = topNKeywords(keywords, o.topN)

	// Top topics
	topicSlice := make([]TopicFreq, 0, len(topicCounts))
	for _, tf := range topicCounts {
		topicSlice = append(topicSlice, *tf)
	}
	sort.Slice(topicSlice, func(i, j int) bool { return topicSlice[i].Count > topicSlice[j].Count })
	if len(topicSlice) > o.topN {
		topicSlice = topicSlice[:o.topN]
	}
	report.TopTopics = topicSlice

	// Peak launch days
	days := make([]DayOfWeek, 0, 7)
	for d := time.Sunday; d <= time.Saturday; d++ {
		if count, ok := dayOfWeek[d]; ok {
			days = append(days, DayOfWeek{Day: d.String(), Count: count})
		}
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Count > days[j].Count })
	report.PeakLaunchDays = days

	// Maker activity
	makers := make([]MakerSummary, 0, len(makerMap))
	for _, ma := range makerMap {
		avg := 0.0
		if ma.Products > 0 {
			avg = float64(ma.TotalVotes) / float64(ma.Products)
		}
		makers = append(makers, MakerSummary{
			UserID: ma.UserID, Username: ma.Username, Name: ma.Name,
			Products: ma.Products, AvgVotes: avg,
		})
	}
	sort.Slice(makers, func(i, j int) bool { return makers[i].Products > makers[j].Products })
	if len(makers) > o.topN {
		makers = makers[:o.topN]
	}
	report.MakerActivity = makers

	return report
}

type makerAccum struct {
	UserID     string
	Username   string
	Name       string
	Products   int
	TotalVotes int
}

func findRisingProducts(posts []Post, topN int) []PostSummary {
	if len(posts) == 0 {
		return nil
	}

	now := time.Now().UTC()
	type scored struct {
		post  Post
		score float64
	}

	var items []scored
	for _, p := range posts {
		age := now.Sub(p.CreatedAt).Hours()
		if age < 1 {
			age = 1
		}
		// Vote velocity: votes per hour
		items = append(items, scored{post: p, score: float64(p.VotesCount) / age})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].score > items[j].score })
	if len(items) > topN {
		items = items[:topN]
	}

	out := make([]PostSummary, 0, len(items))
	for _, it := range items {
		out = append(out, PostSummary{
			ID: it.post.ID, Name: it.post.Name, Tagline: it.post.Tagline,
			Votes: it.post.VotesCount, Comments: it.post.CommentsCount,
			DailyRank: it.post.DailyRank, URL: it.post.URL, Website: it.post.Website,
		})
	}
	return out
}

func extractKeywords(text string, counts map[string]int, stopSet map[string]bool) {
	words := tokenize(text)
	for _, w := range words {
		if len(w) < 3 || stopSet[w] {
			continue
		}
		counts[w]++
	}
	// bigrams
	for i := 0; i < len(words)-1; i++ {
		if stopSet[words[i]] || stopSet[words[i+1]] {
			continue
		}
		bigram := words[i] + " " + words[i+1]
		if len(bigram) >= 5 {
			counts[bigram]++
		}
	}
}

func tokenize(text string) []string {
	lower := strings.ToLower(text)
	var words []string
	var buf strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				words = append(words, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 {
		words = append(words, buf.String())
	}
	return words
}

func topNKeywords(counts map[string]int, n int) []KeywordFreq {
	items := make([]KeywordFreq, 0, len(counts))
	for term, count := range counts {
		items = append(items, KeywordFreq{Term: term, Count: count})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}

func buildStopSet(extra []string) map[string]bool {
	set := make(map[string]bool, len(defaultStopWords)+len(extra))
	for _, w := range defaultStopWords {
		set[w] = true
	}
	for _, w := range extra {
		set[strings.ToLower(w)] = true
	}
	return set
}

var defaultStopWords = []string{
	"a", "an", "and", "are", "as", "at", "be", "been", "being", "but", "by",
	"can", "could", "did", "do", "does", "doing", "done", "each", "even",
	"for", "from", "get", "got", "had", "has", "have", "having", "he", "her",
	"here", "hers", "herself", "him", "himself", "his", "how", "i", "if", "in",
	"into", "is", "it", "its", "itself", "just", "let", "like", "make", "may",
	"me", "might", "more", "most", "much", "must", "my", "myself", "no", "nor",
	"not", "now", "of", "off", "on", "once", "only", "or", "other", "our",
	"ours", "ourselves", "out", "over", "own", "same", "she", "should", "so",
	"some", "still", "such", "than", "that", "the", "their", "theirs", "them",
	"themselves", "then", "there", "these", "they", "this", "those", "through",
	"to", "too", "under", "until", "up", "us", "very", "was", "we", "were",
	"what", "when", "where", "which", "while", "who", "whom", "why", "will",
	"with", "would", "you", "your", "yours", "yourself", "yourselves",
	"about", "above", "after", "again", "against", "all", "also", "am", "any",
	"because", "before", "below", "between", "both", "came", "come", "could",
	"day", "down", "during", "few", "first", "found", "go", "going", "gone",
	"good", "great", "help", "high", "however", "keep", "know", "last", "long",
	"look", "made", "many", "new", "next", "old", "one", "open", "part",
	"place", "point", "put", "right", "said", "say", "see", "set", "since",
	"small", "start", "state", "take", "tell", "think", "thought", "three",
	"time", "turn", "two", "use", "used", "using", "want", "way", "well",
	"went", "work", "world", "year", "yet",
}
