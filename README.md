# producthunt-go

Go client for the Product Hunt GraphQL API. Zero external dependencies (stdlib only).

Used by the GTM agent for product-market fit research, trend analysis,
user feedback mining, and competitive intelligence.

## Quick Start

```go
import producthunt "github.com/teslashibe/producthunt-go"

// Option 1: Developer token (BYOK — full user context, never expires)
c, err := producthunt.New(producthunt.Credentials{
    DeveloperToken: "your-token",
})

// Option 2: Client credentials (auto-provisions token, public scope)
c, err := producthunt.New(producthunt.Credentials{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
})

// Option 3: Browser cookies (frontend API, needs Cloudflare bypass)
c, err := producthunt.New(producthunt.Credentials{
    Session:   "cookie-value",
    CSRFToken: "csrf-value",
})
```

Get credentials at https://www.producthunt.com/v2/oauth/applications.

## Usage

```go
ctx := context.Background()

// Today's top products
feed, _ := c.GetHomefeed(ctx)
for _, p := range feed.Items {
    fmt.Printf("#%d %s — %s (%d votes)\n", p.DailyRank, p.Name, p.Tagline, p.VotesCount)
}

// Search by topic
posts, _ := c.GetTopicPosts(ctx, "artificial-intelligence")

// User profiles & social graph
user, _ := c.GetUser(ctx, "rrhoover")
followers, _ := c.GetUserFollowers(ctx, user.ID)

// Trend analysis
report, _ := c.AnalyzeTrends(ctx,
    producthunt.WithTrendDateRange(from, to),
    producthunt.WithTrendTopics([]string{"developer-tools"}),
)
```

## API Surface

### Posts
- `GetHomefeed` — today's featured products
- `GetPostsByDate` — products by specific date
- `GetTrendingPosts` — currently trending
- `GetNewPosts` — newest (chronological)
- `GetLaunches` — launches by date
- `GetPost` — single post by slug or ID

### Comments
- `GetPostComments` — threaded comments on a post
- `CreateComment` / `ReplyToComment` — engage with products
- `UpdateComment` / `DeleteComment` — manage own comments

### Votes
- `GetPostVoters` — who upvoted a post
- `Upvote` / `RemoveUpvote` — vote management

### Reviews
- `GetPostReviews` — product reviews with ratings
- `CreateReview` — write a review (1.0–5.0 rating)

### Topics
- `GetTopics` — browse/search all topics
- `GetTopic` — single topic by slug
- `GetTopicPosts` — products in a topic
- `FollowTopic` / `UnfollowTopic`

### Collections
- `GetCollections` / `GetCollection` — browse curated lists
- `GetCollectionPosts` — products in a collection
- `CreateCollection` — create a new collection
- `AddPostToCollection` / `RemovePostFromCollection`
- `FollowCollection` / `UnfollowCollection`

### Users
- `GetUser` — profile by username or ID
- `GetUserPosts` / `GetUserVotedPosts` — user activity
- `GetUserFollowers` / `GetUserFollowing` — social graph
- `FollowUser` / `UnfollowUser`

### Search
- `Search` — find posts by keyword

### Viewer
- `GetViewer` — authenticated user profile

### Trends
- `AnalyzeTrends` — aggregate analytics: top keywords, rising products, peak launch days, maker activity

## Probe

```bash
# Developer token
PH_DEV_TOKEN=xxx go run ./cmd/probe

# Client credentials
PH_CLIENT_ID=xxx PH_CLIENT_SECRET=xxx go run ./cmd/probe
```

## Auth Modes

| Mode | Credentials | Scope | Cloudflare |
|---|---|---|---|
| Developer Token | `DeveloperToken` | Full user context | No issues |
| Client Credentials | `ClientID` + `ClientSecret` | Public (no viewer) | No issues |
| Browser Cookies | `Session` + `CSRFToken` | Full + writes | Requires bypass |

The v2 API at `api.producthunt.com` has no Cloudflare challenge.
The frontend API at `www.producthunt.com/frontend/graphql` requires
Cloudflare managed challenge bypass (TLS fingerprint spoofing).
