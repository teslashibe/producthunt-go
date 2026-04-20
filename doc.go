// Package producthunt is a Go client for Product Hunt's GraphQL API.
//
// Three authentication modes are supported (in priority order):
//
// 1. Developer token (BYOK — full user context, never expires):
//
//	c, err := producthunt.New(producthunt.Credentials{
//	    DeveloperToken: "your-token",
//	})
//
// 2. Client credentials (auto-provisioned, public scope):
//
//	c, err := producthunt.New(producthunt.Credentials{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	})
//
// 3. Browser cookies (frontend API, requires Cloudflare bypass):
//
//	c, err := producthunt.New(producthunt.Credentials{
//	    Session:   "HgMBNVq0dc44...",
//	    CSRFToken: "WGqV7WPA...",
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
