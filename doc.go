// Package producthunt is a Go client for Product Hunt's GraphQL API.
//
// Two authentication modes are supported:
//
// Developer token (recommended for reads):
//
//	c, err := producthunt.New(producthunt.Cookies{
//	    DeveloperToken: "your-token-from-producthunt.com/v2/oauth/applications",
//	})
//
// Browser cookies (required for writes — upvote, comment, follow):
//
//	c, err := producthunt.New(producthunt.Cookies{
//	    Session:     "HgMBNVq0dc44...",
//	    CFClearance: "abc123...",
//	    CSRFToken:   "WGqV7WPA...",
//	})
//
// Example:
//
//	feed, err := c.GetHomefeed(ctx)
//	for _, post := range feed.Items {
//	    fmt.Printf("%s — %s (%d votes)\n", post.Name, post.Tagline, post.VotesCount)
//	}
//
// The client covers the full programmatic surface of the site:
//   - Posts: homefeed, by-date, trending, new, launches, single post lookup
//   - Comments: read threads, create, reply, update, delete
//   - Votes: upvote, remove upvote, list voters
//   - Reviews: read reviews, create review
//   - Topics: browse, search, follow/unfollow, posts-by-topic
//   - Collections: browse, create, manage, follow/unfollow
//   - Users: profiles, social graph, follow/unfollow
//   - Search: full-text across posts, users, collections
//   - Trends: aggregate analytics for product-market fit research
package producthunt
