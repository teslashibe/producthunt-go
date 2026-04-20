// probe is a quick integration test for the producthunt-go client.
//
// Usage:
//
//	# With developer token (recommended):
//	PH_DEV_TOKEN=your-token go run ./cmd/probe
//
//	# With browser cookies (needs Cloudflare bypass):
//	PH_SESSION=cookie-value PH_CF_CLEARANCE=cookie-value go run ./cmd/probe
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	producthunt "github.com/teslashibe/producthunt-go"
)

func main() {
	cookies := producthunt.Cookies{
		DeveloperToken: os.Getenv("PH_DEV_TOKEN"),
		Session:        os.Getenv("PH_SESSION"),
		CFClearance:    os.Getenv("PH_CF_CLEARANCE"),
		CFBM:           os.Getenv("PH_CF_BM"),
		CSRFToken:      os.Getenv("PH_CSRF_TOKEN"),
	}

	if cookies.DeveloperToken == "" && cookies.Session == "" {
		fmt.Fprintln(os.Stderr, "Set PH_DEV_TOKEN or PH_SESSION env var")
		fmt.Fprintln(os.Stderr, "Get a developer token at: https://www.producthunt.com/v2/oauth/applications")
		os.Exit(1)
	}

	fmt.Println("=== Creating client...")
	c, err := producthunt.New(cookies, producthunt.WithRateLimit(500*time.Millisecond))
	if err != nil {
		fmt.Fprintf(os.Stderr, "New() failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Client created successfully")

	ctx := context.Background()

	fmt.Println("\n=== GetViewer")
	viewer, err := c.GetViewer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetViewer: %v\n", err)
	} else {
		fmt.Printf("  ID:       %s\n", viewer.ID)
		fmt.Printf("  Username: %s\n", viewer.Username)
		fmt.Printf("  Name:     %s\n", viewer.Name)
		fmt.Printf("  Headline: %s\n", viewer.Headline)
	}

	fmt.Println("\n=== GetHomefeed (top 5)")
	feed, err := c.GetHomefeed(ctx, producthunt.WithPostListFirst(5))
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetHomefeed: %v\n", err)
	} else {
		fmt.Printf("  Total: %d, HasNext: %v\n", feed.TotalCount, feed.HasNext)
		for i, p := range feed.Items {
			fmt.Printf("  %d. %s — %s (%d votes, rank #%d)\n", i+1, p.Name, p.Tagline, p.VotesCount, p.DailyRank)
		}
	}

	fmt.Println("\n=== GetTopics (top 5)")
	topics, err := c.GetTopics(ctx, producthunt.WithTopicsOrder(producthunt.TopicsOrderFollowers))
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetTopics: %v\n", err)
	} else {
		for i, t := range topics.Items {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s (%d posts, %d followers)\n", i+1, t.Name, t.PostsCount, t.FollowersCount)
		}
	}

	fmt.Println("\n=== Done")
}
