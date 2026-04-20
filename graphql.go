package producthunt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// gqlRequest is the JSON body sent to the frontend GraphQL endpoint.
type gqlRequest struct {
	OperationName string      `json:"operationName"`
	Variables     interface{} `json:"variables"`
	Query         string      `json:"query"`
}

// gqlResponse is the top-level GraphQL response envelope.
type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors"`
}

type gqlError struct {
	Message   string   `json:"message"`
	Path      []string `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
}

func (r *gqlResponse) err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	first := r.Errors[0]
	msg := strings.ToLower(first.Message)

	switch {
	case strings.Contains(msg, "not authenticated"),
		strings.Contains(msg, "unauthorized"),
		strings.Contains(msg, "session expired"),
		strings.Contains(msg, "login required"):
		return ErrUnauthorized

	case strings.Contains(msg, "already voted"),
		strings.Contains(msg, "already upvoted"):
		return ErrAlreadyVoted

	case strings.Contains(msg, "not voted"),
		strings.Contains(msg, "have not voted"):
		return ErrNotVoted

	case strings.Contains(msg, "not found"),
		strings.Contains(msg, "does not exist"),
		strings.Contains(msg, "couldn't find"):
		return ErrNotFound

	case strings.Contains(msg, "forbidden"),
		strings.Contains(msg, "not allowed"),
		strings.Contains(msg, "access denied"),
		strings.Contains(msg, "permission"):
		return ErrForbidden

	case strings.Contains(msg, "rate limit"),
		strings.Contains(msg, "throttle"):
		return ErrRateLimited
	}

	return fmt.Errorf("%w: %s", ErrGraphQL, first.Message)
}

// unmarshalData decodes raw GraphQL response data into v.
func unmarshalData(raw json.RawMessage, v interface{}) error {
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%w: decoding response: %v (snippet: %s)",
			ErrRequestFailed, err, truncate(string(raw), 300))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// --- Relay connection helpers ---

type relayPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type relayConnection[T any] struct {
	TotalCount int           `json:"totalCount"`
	PageInfo   relayPageInfo `json:"pageInfo"`
	Edges      []struct {
		Cursor string `json:"cursor"`
		Node   T      `json:"node"`
	} `json:"edges"`
}

func (c *relayConnection[T]) toPage() Page[T] {
	p := Page[T]{
		TotalCount: c.TotalCount,
		HasNext:    c.PageInfo.HasNextPage,
		EndCursor:  c.PageInfo.EndCursor,
	}
	p.Items = make([]T, 0, len(c.Edges))
	for _, e := range c.Edges {
		p.Items = append(p.Items, e.Node)
	}
	return p
}

// --- Internal GraphQL response shapes ---

type gqlPost struct {
	ID            string    `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Tagline       string    `json:"tagline"`
	Description   string    `json:"description"`
	URL           string    `json:"url"`
	Website       string    `json:"website"`
	VotesCount    int       `json:"votesCount"`
	CommentsCount int       `json:"commentsCount"`
	ReviewsCount  int       `json:"reviewsCount"`
	ReviewsRating float64   `json:"reviewsRating"`
	DailyRank     *int      `json:"dailyRank"`
	WeeklyRank    *int      `json:"weeklyRank"`
	MonthlyRank   *int      `json:"monthlyRank"`
	YearlyRank    *int      `json:"yearlyRank"`
	FeaturedAt    *string   `json:"featuredAt"`
	CreatedAt     string    `json:"createdAt"`
	ScheduledAt   *string   `json:"scheduledAt"`
	IsVoted       bool      `json:"isVoted"`
	IsCollected   bool      `json:"isCollected"`
	Thumbnail     *gqlMedia `json:"thumbnail"`
	Media         []gqlMedia      `json:"media"`
	ProductLinks  []gqlProductLink `json:"productLinks"`
	Topics        *relayConnection[gqlTopic] `json:"topics"`
	Makers        []gqlUser     `json:"makers"`
	User          gqlUser       `json:"user"`
}

func (p *gqlPost) toPost() Post {
	out := Post{
		ID:            p.ID,
		Slug:          p.Slug,
		Name:          p.Name,
		Tagline:       p.Tagline,
		Description:   p.Description,
		URL:           p.URL,
		Website:       p.Website,
		VotesCount:    p.VotesCount,
		CommentsCount: p.CommentsCount,
		ReviewsCount:  p.ReviewsCount,
		ReviewsRating: p.ReviewsRating,
		IsVoted:       p.IsVoted,
		IsCollected:   p.IsCollected,
		CreatedAt:     parseTime(p.CreatedAt),
		User:          p.User.toUser(),
	}
	if p.DailyRank != nil {
		out.DailyRank = *p.DailyRank
	}
	if p.WeeklyRank != nil {
		out.WeeklyRank = *p.WeeklyRank
	}
	if p.MonthlyRank != nil {
		out.MonthlyRank = *p.MonthlyRank
	}
	if p.YearlyRank != nil {
		out.YearlyRank = *p.YearlyRank
	}
	if p.FeaturedAt != nil {
		out.FeaturedAt = parseTime(*p.FeaturedAt)
	}
	if p.ScheduledAt != nil {
		out.ScheduledAt = parseTime(*p.ScheduledAt)
	}
	if p.Thumbnail != nil {
		m := p.Thumbnail.toMedia()
		out.Thumbnail = &m
	}
	for _, m := range p.Media {
		out.Media = append(out.Media, m.toMedia())
	}
	for _, l := range p.ProductLinks {
		out.ProductLinks = append(out.ProductLinks, ProductLink{Type: l.Type, URL: l.URL})
	}
	if p.Topics != nil {
		for _, e := range p.Topics.Edges {
			out.Topics = append(out.Topics, e.Node.toTopic())
		}
	}
	for _, m := range p.Makers {
		out.Makers = append(out.Makers, m.toUser())
	}
	return out
}

type gqlUser struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	Headline       string  `json:"headline"`
	ProfileImage   string  `json:"profileImage"`
	CoverImage     string  `json:"coverImage"`
	TwitterUsername string  `json:"twitterUsername"`
	WebsiteURL     string  `json:"websiteUrl"`
	URL            string  `json:"url"`
	FollowersCount int     `json:"followersCount"`
	FollowingCount int     `json:"followingCount"`
	IsMaker        bool    `json:"isMaker"`
	IsFollowing    bool    `json:"isFollowing"`
	IsViewer       bool    `json:"isViewer"`
	CreatedAt      *string `json:"createdAt"`
}

func (u *gqlUser) toUser() User {
	out := User{
		ID:             u.ID,
		Username:       u.Username,
		Name:           u.Name,
		Headline:       u.Headline,
		ProfileImage:   u.ProfileImage,
		CoverImage:     u.CoverImage,
		TwitterUsername: u.TwitterUsername,
		WebsiteURL:     u.WebsiteURL,
		URL:            u.URL,
		FollowersCount: u.FollowersCount,
		FollowingCount: u.FollowingCount,
		IsMaker:        u.IsMaker,
		IsFollowing:    u.IsFollowing,
		IsViewer:       u.IsViewer,
	}
	if u.CreatedAt != nil {
		out.CreatedAt = parseTime(*u.CreatedAt)
	}
	return out
}

type gqlComment struct {
	ID         string    `json:"id"`
	Body       string    `json:"body"`
	VotesCount int       `json:"votesCount"`
	IsVoted    bool      `json:"isVoted"`
	User       gqlUser   `json:"user"`
	ParentID   *string   `json:"parentId"`
	CreatedAt  string    `json:"createdAt"`
	Replies    *relayConnection[gqlComment] `json:"replies"`
}

func (c *gqlComment) toComment() Comment {
	out := Comment{
		ID:         c.ID,
		Body:       c.Body,
		VotesCount: c.VotesCount,
		IsVoted:    c.IsVoted,
		User:       c.User.toUser(),
		CreatedAt:  parseTime(c.CreatedAt),
	}
	if c.ParentID != nil {
		out.ParentID = *c.ParentID
	}
	if c.Replies != nil {
		for _, e := range c.Replies.Edges {
			out.Replies = append(out.Replies, e.Node.toComment())
		}
	}
	return out
}

type gqlTopic struct {
	ID             string  `json:"id"`
	Slug           string  `json:"slug"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Image          string  `json:"image"`
	PostsCount     int     `json:"postsCount"`
	FollowersCount int     `json:"followersCount"`
	IsFollowing    bool    `json:"isFollowing"`
	URL            string  `json:"url"`
	CreatedAt      *string `json:"createdAt"`
}

func (t *gqlTopic) toTopic() Topic {
	out := Topic{
		ID:             t.ID,
		Slug:           t.Slug,
		Name:           t.Name,
		Description:    t.Description,
		Image:          t.Image,
		PostsCount:     t.PostsCount,
		FollowersCount: t.FollowersCount,
		IsFollowing:    t.IsFollowing,
		URL:            t.URL,
	}
	if t.CreatedAt != nil {
		out.CreatedAt = parseTime(*t.CreatedAt)
	}
	return out
}

type gqlCollection struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Tagline        string  `json:"tagline"`
	Description    string  `json:"description"`
	CoverImage     string  `json:"coverImage"`
	FollowersCount int     `json:"followersCount"`
	IsFollowing    bool    `json:"isFollowing"`
	URL            string  `json:"url"`
	User           gqlUser `json:"user"`
	FeaturedAt     *string `json:"featuredAt"`
	CreatedAt      string  `json:"createdAt"`
}

func (c *gqlCollection) toCollection() Collection {
	out := Collection{
		ID:             c.ID,
		Name:           c.Name,
		Tagline:        c.Tagline,
		Description:    c.Description,
		CoverImage:     c.CoverImage,
		FollowersCount: c.FollowersCount,
		IsFollowing:    c.IsFollowing,
		URL:            c.URL,
		User:           c.User.toUser(),
		CreatedAt:      parseTime(c.CreatedAt),
	}
	if c.FeaturedAt != nil {
		out.FeaturedAt = parseTime(*c.FeaturedAt)
	}
	return out
}

type gqlVote struct {
	ID        string  `json:"id"`
	User      gqlUser `json:"user"`
	CreatedAt string  `json:"createdAt"`
}

func (v *gqlVote) toVote() Vote {
	return Vote{
		ID:        v.ID,
		User:      v.User.toUser(),
		CreatedAt: parseTime(v.CreatedAt),
	}
}

type gqlReview struct {
	ID        string  `json:"id"`
	Body      string  `json:"body"`
	Rating    float64 `json:"rating"`
	Sentiment string  `json:"sentiment"`
	User      gqlUser `json:"user"`
	CreatedAt string  `json:"createdAt"`
}

func (r *gqlReview) toReview() Review {
	return Review{
		ID:        r.ID,
		Body:      r.Body,
		Rating:    r.Rating,
		Sentiment: r.Sentiment,
		User:      r.User.toUser(),
		CreatedAt: parseTime(r.CreatedAt),
	}
}

type gqlMedia struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	VideoURL string `json:"videoUrl"`
}

func (m *gqlMedia) toMedia() Media {
	return Media{Type: m.Type, URL: m.URL, VideoURL: m.VideoURL}
}

type gqlProductLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// parseTime parses an ISO 8601 timestamp from the API.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, _ = time.Parse("2006-01-02T15:04:05Z", s)
	}
	return t
}
