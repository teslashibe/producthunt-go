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

## MCP support

This package ships an [MCP](https://modelcontextprotocol.io/) tool surface in
`./mcp` for use with [`teslashibe/mcptool`](https://github.com/teslashibe/mcptool)-compatible
hosts (e.g. [`teslashibe/agent-setup`](https://github.com/teslashibe/agent-setup)).
39 tools cover the full client API: home/trending/new/by-date feeds, single
post fetch, post comments/voters/reviews, vote and review writes, user
profiles + social graph + follow writes, viewer self-fetch, topic browse +
follow, collection browse + create + add/remove/follow, search, and the
aggregate `producthunt_analyze_trends` tool.

```go
import (
    "github.com/teslashibe/mcptool"
    producthunt "github.com/teslashibe/producthunt-go"
    phmcp "github.com/teslashibe/producthunt-go/mcp"
)

client, _ := producthunt.New(producthunt.Credentials{...})
provider := phmcp.Provider{}
for _, tool := range provider.Tools() {
    // register tool with your MCP server, passing client as the
    // opaque client argument when invoking
}
```

A coverage test in `mcp/mcp_test.go` fails if a new exported method is added
to `*Client` without either being wrapped by an MCP tool or being added to
`mcp.Excluded` with a reason — keeping the MCP surface in lockstep with the
package API is enforced by CI rather than convention.

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
