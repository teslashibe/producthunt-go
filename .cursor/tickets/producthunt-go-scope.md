# `producthunt-go` — Full API Scope

**Repo:** `github.com/teslashibe/producthunt-go`  
**Package:** `github.com/teslashibe/producthunt-go`  
**Mirrors:** `hn-go` / `facebook-go` / `reddit-go` conventions (stdlib only, zero prod deps)  
**Purpose:** Authenticated Go client for Product Hunt's internal frontend GraphQL API, used by the GTM agent for product-market fit research, trend analysis, user feedback mining, and competitive intelligence.

---

## Background & Reverse-Engineering Notes

Product Hunt exposes two GraphQL APIs:

1. **Public API v2** — `https://api.producthunt.com/v2/api/graphql` (OAuth2 tokens, rate-limited, read-heavy, limited mutations)
2. **Internal Frontend API** — `https://www.producthunt.com/frontend/graphql` (cookie-based, used by the website itself, full read/write access)

This client targets the **internal frontend API** for maximum capability — upvoting,
commenting, launching products, searching, following — all the actions the website
itself performs. The public API v2 schema is used as the baseline reference for
the data model, with internal extensions discovered via browser DevTools network
inspection.

### Cloudflare Challenge

Product Hunt sits behind Cloudflare with browser challenge protection. Plain `curl`
without the right `cf_clearance` cookie gets a 403. The client must carry Cloudflare
clearance cookies from the browser session alongside the PH session cookies.

### Required Cookies (from browser session export)

| Cookie | Domain | httpOnly | Purpose |
|---|---|---|---|
| `_producthunt_session_production` | `.producthunt.com` | ✓ | **Rails session** — primary auth cookie |
| `cf_clearance` | `.producthunt.com` | — | **Cloudflare challenge clearance** |
| `__cf_bm` | `.producthunt.com` | ✓ | Cloudflare bot management |
| `_ph_id` | `.producthunt.com` | — | PH analytics / visitor ID |
| `ph_locale` | `.producthunt.com` | — | Locale preference |

> **Note:** The exact cookie set needs verification with an actual PH browser session
> export. The `_producthunt_session_production` cookie is the critical auth credential;
> the Cloudflare cookies (`cf_clearance`, `__cf_bm`) are required to pass bot detection.

### GraphQL Request Structure

```
POST https://www.producthunt.com/frontend/graphql
Content-Type: application/json
Accept: application/json

Headers (required):
  User-Agent:      Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) ...Chrome/131...
  Referer:         https://www.producthunt.com/
  Origin:          https://www.producthunt.com
  Accept:          application/json
  Accept-Language: en-US,en;q=0.9
  X-Requested-With: XMLHttpRequest
  Cookie:          _producthunt_session_production=<session>; cf_clearance=<cf>; ...

Body (JSON):
  {
    "operationName": "<QueryName>",
    "variables": { ... },
    "query": "query QueryName($var: Type!) { ... }"
  }
```

### Response Format

Standard GraphQL JSON response:

```json
{
  "data": { ... },
  "errors": [{ "message": "...", "locations": [...], "path": [...] }]
}
```

Pagination uses Relay-style cursor connections: `{ edges: [{ node, cursor }], pageInfo: { hasNextPage, endCursor } }`.

---

## API Surface: GraphQL Operations

### Queries (Read)

The following operations map to Product Hunt's documented schema plus internal
frontend-only extensions discovered via network inspection.

| # | Operation | Description | Key Variables |
|---|---|---|---|
| Q1 | `HomefeedQuery` | Today's featured/hot products (the main feed) | `date`, `cursor`, `first` |
| Q2 | `PostsByDateQuery` | Posts filtered by date (daily leaderboard) | `postedAfter`, `postedBefore`, `featured`, `order`, `first`, `after` |
| Q3 | `PostQuery` | Single post by ID or slug | `id` or `slug` |
| Q4 | `PostCommentsQuery` | Comments on a post with threading | `postId`, `order`, `first`, `after` |
| Q5 | `PostVotersQuery` | List of users who upvoted a post | `postId`, `first`, `after` |
| Q6 | `PostReviewsQuery` | Reviews (distinct from comments) on a post | `postId`, `first`, `after` |
| Q7 | `TopicsQuery` | Browse all topics with pagination | `query`, `order`, `first`, `after`, `followedByUserid` |
| Q8 | `TopicQuery` | Single topic by slug | `slug` |
| Q9 | `TopicPostsQuery` | Posts under a specific topic | `topicSlug`, `order`, `first`, `after` |
| Q10 | `CollectionsQuery` | Browse collections | `featured`, `userId`, `postId`, `order`, `first`, `after` |
| Q11 | `CollectionQuery` | Single collection by ID | `id` |
| Q12 | `CollectionPostsQuery` | Posts within a collection | `collectionId`, `first`, `after` |
| Q13 | `UserQuery` | User profile by ID or username | `id` or `username` |
| Q14 | `UserPostsQuery` | Posts made by a user | `userId`, `first`, `after` |
| Q15 | `UserVotedPostsQuery` | Posts upvoted by a user | `userId`, `first`, `after` |
| Q16 | `UserFollowersQuery` | Users following a user | `userId`, `first`, `after` |
| Q17 | `UserFollowingQuery` | Users a user follows | `userId`, `first`, `after` |
| Q18 | `SearchQuery` | Full-text search across posts, collections, users | `query`, `type`, `first`, `after` |
| Q19 | `ViewerQuery` | Currently authenticated user | (none) |
| Q20 | `TrendingPostsQuery` | Posts trending right now (time-weighted) | `first`, `after` |
| Q21 | `NewPostsQuery` | Newest posts (unranked, chronological) | `first`, `after` |
| Q22 | `LaunchesQuery` | Products scheduled to launch / recently launched | `date`, `first`, `after` |
| Q23 | `NewsletterQuery` | Daily digest / newsletter content | `date` |

### Mutations (Write)

| # | Operation | Description | Key Variables |
|---|---|---|---|
| M1 | `VoteCreateMutation` | Upvote a post | `postId` |
| M2 | `VoteDeleteMutation` | Remove upvote from a post | `postId` |
| M3 | `CommentCreateMutation` | Post a comment on a product | `postId`, `body`, `parentId` (for replies) |
| M4 | `CommentUpdateMutation` | Edit own comment | `commentId`, `body` |
| M5 | `CommentDeleteMutation` | Delete own comment | `commentId` |
| M6 | `UserFollowMutation` | Follow a user | `userId` |
| M7 | `UserUnfollowMutation` | Unfollow a user | `userId` |
| M8 | `CollectionFollowMutation` | Follow a collection | `collectionId` |
| M9 | `CollectionUnfollowMutation` | Unfollow a collection | `collectionId` |
| M10 | `TopicFollowMutation` | Follow a topic | `topicId` |
| M11 | `TopicUnfollowMutation` | Unfollow a topic | `topicId` |
| M12 | `CollectionAddPostMutation` | Add a post to own collection | `collectionId`, `postId` |
| M13 | `CollectionRemovePostMutation` | Remove a post from own collection | `collectionId`, `postId` |
| M14 | `CollectionCreateMutation` | Create a new collection | `name`, `tagline`, `description` |
| M15 | `ReviewCreateMutation` | Write a review for a product | `postId`, `body`, `rating`, `sentiment` |

> **Note:** Operation names above are best-guesses based on the public schema and
> standard PH frontend naming patterns. The exact `operationName` strings will be
> verified by inspecting browser DevTools Network tab with an authenticated session.
> Some operations may use numbered `doc_id` patterns similar to Facebook's internal API.

---

## File Structure

```
producthunt-go/                     ← repo root
├── go.mod                          # module github.com/teslashibe/producthunt-go
├── README.md
├── .cursor/
│   └── tickets/
│       └── producthunt-go-scope.md # this file
├── doc.go                          # package producthunt — top-level doc
├── client.go                       # Client, New(), Option funcs, Cookies type, HTTP layer
├── posts.go                        # GetHomefeed, GetPostsByDate, GetPost, GetTrendingPosts,
│                                   #   GetNewPosts, GetLaunches
├── comments.go                     # GetPostComments, CreateComment, UpdateComment, DeleteComment
├── votes.go                        # GetPostVoters, Upvote, RemoveUpvote
├── reviews.go                      # GetPostReviews, CreateReview
├── topics.go                       # GetTopics, GetTopic, GetTopicPosts,
│                                   #   FollowTopic, UnfollowTopic
├── collections.go                  # GetCollections, GetCollection, GetCollectionPosts,
│                                   #   CreateCollection, FollowCollection, UnfollowCollection,
│                                   #   AddPostToCollection, RemovePostFromCollection
├── users.go                        # GetUser, GetUserPosts, GetUserVotedPosts,
│                                   #   GetUserFollowers, GetUserFollowing,
│                                   #   FollowUser, UnfollowUser
├── search.go                       # Search (posts, users, collections)
├── viewer.go                       # GetViewer (authenticated user)
├── trends.go                       # AnalyzeTrends → TrendReport (aggregate analytics)
├── types.go                        # All exported domain types
├── errors.go                       # Sentinel errors
└── graphql.go                      # Internal: query builder, response parsing, pagination helpers
```

---

## Domain Types (`types.go`)

```go
// Cookies holds the Product Hunt session cookies from a browser export.
type Cookies struct {
    Session     string // _producthunt_session_production
    CFClearance string // cf_clearance
    CFBM        string // __cf_bm
    PHID        string // _ph_id (optional, analytics)
}

// Post is a product launched on Product Hunt.
type Post struct {
    ID            string    `json:"id"`
    Slug          string    `json:"slug"`
    Name          string    `json:"name"`
    Tagline       string    `json:"tagline"`
    Description   string    `json:"description,omitempty"`
    URL           string    `json:"url"`
    Website       string    `json:"website"`
    VotesCount    int       `json:"votesCount"`
    CommentsCount int       `json:"commentsCount"`
    ReviewsCount  int       `json:"reviewsCount"`
    ReviewsRating float64   `json:"reviewsRating"`
    DailyRank     int       `json:"dailyRank,omitempty"`
    WeeklyRank    int       `json:"weeklyRank,omitempty"`
    MonthlyRank   int       `json:"monthlyRank,omitempty"`
    YearlyRank    int       `json:"yearlyRank,omitempty"`
    FeaturedAt    time.Time `json:"featuredAt,omitempty"`
    CreatedAt     time.Time `json:"createdAt"`
    ScheduledAt   time.Time `json:"scheduledAt,omitempty"`
    Thumbnail     *Media    `json:"thumbnail,omitempty"`
    Media         []Media   `json:"media,omitempty"`
    ProductLinks  []ProductLink `json:"productLinks,omitempty"`
    Topics        []Topic   `json:"topics,omitempty"`
    Makers        []User    `json:"makers,omitempty"`
    User          User      `json:"user"`
    IsVoted       bool      `json:"isVoted"`
    IsCollected   bool      `json:"isCollected"`
}

// User is a Product Hunt user.
type User struct {
    ID              string `json:"id"`
    Username        string `json:"username"`
    Name            string `json:"name"`
    Headline        string `json:"headline,omitempty"`
    ProfileImage    string `json:"profileImage,omitempty"`
    CoverImage      string `json:"coverImage,omitempty"`
    TwitterUsername  string `json:"twitterUsername,omitempty"`
    WebsiteURL      string `json:"websiteUrl,omitempty"`
    URL             string `json:"url"`
    FollowersCount  int    `json:"followersCount"`
    FollowingCount  int    `json:"followingCount"`
    IsMaker         bool   `json:"isMaker"`
    IsFollowing     bool   `json:"isFollowing"`
    IsViewer        bool   `json:"isViewer"`
    CreatedAt       time.Time `json:"createdAt,omitempty"`
}

// Comment is a comment on a post.
type Comment struct {
    ID         string    `json:"id"`
    Body       string    `json:"body"`
    VotesCount int       `json:"votesCount"`
    IsVoted    bool      `json:"isVoted"`
    User       User      `json:"user"`
    ParentID   string    `json:"parentId,omitempty"`
    CreatedAt  time.Time `json:"createdAt"`
    Replies    []Comment `json:"replies,omitempty"`
}

// Topic is a product category.
type Topic struct {
    ID             string `json:"id"`
    Slug           string `json:"slug"`
    Name           string `json:"name"`
    Description    string `json:"description,omitempty"`
    Image          string `json:"image,omitempty"`
    PostsCount     int    `json:"postsCount"`
    FollowersCount int    `json:"followersCount"`
    IsFollowing    bool   `json:"isFollowing"`
    URL            string `json:"url"`
    CreatedAt      time.Time `json:"createdAt,omitempty"`
}

// Collection is a curated list of posts.
type Collection struct {
    ID             string    `json:"id"`
    Name           string    `json:"name"`
    Tagline        string    `json:"tagline"`
    Description    string    `json:"description,omitempty"`
    CoverImage     string    `json:"coverImage,omitempty"`
    FollowersCount int       `json:"followersCount"`
    IsFollowing    bool      `json:"isFollowing"`
    URL            string    `json:"url"`
    User           User      `json:"user"`
    FeaturedAt     time.Time `json:"featuredAt,omitempty"`
    CreatedAt      time.Time `json:"createdAt"`
}

// Vote represents an upvote on a post.
type Vote struct {
    ID        string    `json:"id"`
    User      User      `json:"user"`
    CreatedAt time.Time `json:"createdAt"`
}

// Review is a user review of a product.
type Review struct {
    ID        string    `json:"id"`
    Body      string    `json:"body"`
    Rating    float64   `json:"rating"`
    Sentiment string    `json:"sentiment,omitempty"` // positive / neutral / negative
    User      User      `json:"user"`
    CreatedAt time.Time `json:"createdAt"`
}

// Media is an image or video associated with a post.
type Media struct {
    Type     string `json:"type"` // "image" | "video"
    URL      string `json:"url"`
    VideoURL string `json:"videoUrl,omitempty"`
}

// ProductLink is an additional link on a post (App Store, Play Store, etc.).
type ProductLink struct {
    Type string `json:"type"`
    URL  string `json:"url"`
}

// SearchType restricts search results.
type SearchType string
const (
    SearchPosts       SearchType = "POSTS"
    SearchUsers       SearchType = "USERS"
    SearchCollections SearchType = "COLLECTIONS"
)

// PostsOrder defines sort order for post queries.
type PostsOrder string
const (
    PostsOrderRanking PostsOrder = "RANKING"
    PostsOrderNewest  PostsOrder = "NEWEST"
    PostsOrderVotes   PostsOrder = "VOTES"
)

// CommentsOrder defines sort order for comment queries.
type CommentsOrder string
const (
    CommentsOrderNewest CommentsOrder = "NEWEST"
    CommentsOrderOldest CommentsOrder = "OLDEST"
)

// TopicsOrder defines sort order for topic queries.
type TopicsOrder string
const (
    TopicsOrderNewest    TopicsOrder = "NEWEST"
    TopicsOrderFollowers TopicsOrder = "FOLLOWERS_COUNT"
)

// CollectionsOrder defines sort order for collection queries.
type CollectionsOrder string
const (
    CollectionsOrderFeatured CollectionsOrder = "FEATURED_AT"
    CollectionsOrderFollowers CollectionsOrder = "FOLLOWERS_COUNT"
    CollectionsOrderNewest   CollectionsOrder = "NEWEST"
)

// Page is a generic paginated response.
type Page[T any] struct {
    Items      []T    `json:"items"`
    TotalCount int    `json:"totalCount"`
    EndCursor  string `json:"endCursor,omitempty"`
    HasNext    bool   `json:"hasNext"`
}

// SearchResult wraps typed search results.
type SearchResult struct {
    Posts       Page[Post]       `json:"posts,omitempty"`
    Users       Page[User]       `json:"users,omitempty"`
    Collections Page[Collection] `json:"collections,omitempty"`
}
```

---

## API Surface (`client.go`, all exported methods on `*Client`)

```go
// Auth
func New(cookies Cookies, opts ...Option) (*Client, error)

// --- Posts (reads) ---
func (c *Client) GetHomefeed(ctx context.Context, opts ...PostListOption) (Page[Post], error)
func (c *Client) GetPostsByDate(ctx context.Context, date time.Time, opts ...PostListOption) (Page[Post], error)
func (c *Client) GetTrendingPosts(ctx context.Context, opts ...PostListOption) (Page[Post], error)
func (c *Client) GetNewPosts(ctx context.Context, opts ...PostListOption) (Page[Post], error)
func (c *Client) GetLaunches(ctx context.Context, date time.Time, opts ...PostListOption) (Page[Post], error)
func (c *Client) GetPost(ctx context.Context, slugOrID string) (*Post, error)

// --- Comments ---
func (c *Client) GetPostComments(ctx context.Context, postID string, opts ...CommentListOption) (Page[Comment], error)
func (c *Client) CreateComment(ctx context.Context, postID, body string) (*Comment, error)
func (c *Client) ReplyToComment(ctx context.Context, postID, parentCommentID, body string) (*Comment, error)
func (c *Client) UpdateComment(ctx context.Context, commentID, body string) (*Comment, error)
func (c *Client) DeleteComment(ctx context.Context, commentID string) error

// --- Votes ---
func (c *Client) GetPostVoters(ctx context.Context, postID string, opts ...PageOption) (Page[Vote], error)
func (c *Client) Upvote(ctx context.Context, postID string) error
func (c *Client) RemoveUpvote(ctx context.Context, postID string) error

// --- Reviews ---
func (c *Client) GetPostReviews(ctx context.Context, postID string, opts ...PageOption) (Page[Review], error)
func (c *Client) CreateReview(ctx context.Context, postID string, body string, rating float64) (*Review, error)

// --- Topics ---
func (c *Client) GetTopics(ctx context.Context, opts ...TopicListOption) (Page[Topic], error)
func (c *Client) GetTopic(ctx context.Context, slug string) (*Topic, error)
func (c *Client) GetTopicPosts(ctx context.Context, topicSlug string, opts ...PostListOption) (Page[Post], error)
func (c *Client) FollowTopic(ctx context.Context, topicID string) error
func (c *Client) UnfollowTopic(ctx context.Context, topicID string) error

// --- Collections ---
func (c *Client) GetCollections(ctx context.Context, opts ...CollectionListOption) (Page[Collection], error)
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*Collection, error)
func (c *Client) GetCollectionPosts(ctx context.Context, collectionID string, opts ...PostListOption) (Page[Post], error)
func (c *Client) CreateCollection(ctx context.Context, params CreateCollectionParams) (*Collection, error)
func (c *Client) AddPostToCollection(ctx context.Context, collectionID, postID string) error
func (c *Client) RemovePostFromCollection(ctx context.Context, collectionID, postID string) error
func (c *Client) FollowCollection(ctx context.Context, collectionID string) error
func (c *Client) UnfollowCollection(ctx context.Context, collectionID string) error

// --- Users ---
func (c *Client) GetUser(ctx context.Context, usernameOrID string) (*User, error)
func (c *Client) GetUserPosts(ctx context.Context, userID string, opts ...PageOption) (Page[Post], error)
func (c *Client) GetUserVotedPosts(ctx context.Context, userID string, opts ...PageOption) (Page[Post], error)
func (c *Client) GetUserFollowers(ctx context.Context, userID string, opts ...PageOption) (Page[User], error)
func (c *Client) GetUserFollowing(ctx context.Context, userID string, opts ...PageOption) (Page[User], error)
func (c *Client) FollowUser(ctx context.Context, userID string) error
func (c *Client) UnfollowUser(ctx context.Context, userID string) error

// --- Search ---
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOption) (*SearchResult, error)

// --- Viewer ---
func (c *Client) GetViewer(ctx context.Context) (*User, error)

// --- Trends (aggregate analytics) ---
func (c *Client) AnalyzeTrends(ctx context.Context, opts ...TrendOption) (*TrendReport, error)
```

### Options & Params

```go
// Client options
func WithUserAgent(ua string) Option
func WithHTTPClient(hc *http.Client) Option
func WithProxy(proxyURL string) Option
func WithRateLimit(d time.Duration) Option       // default 1s between requests
func WithRetry(maxAttempts int, base time.Duration) Option

// Pagination options (shared across list queries)
type PageOption func(*pageOptions)
func WithFirst(n int) PageOption                  // default 20, max 50
func WithAfter(cursor string) PageOption

// Post list options (extends PageOption)
type PostListOption func(*postListOptions)
func WithPostsOrder(order PostsOrder) PostListOption
func WithFeatured(featured bool) PostListOption
func WithPostedAfter(t time.Time) PostListOption
func WithPostedBefore(t time.Time) PostListOption
func WithPostsTopic(slug string) PostListOption

// Comment list options
type CommentListOption func(*commentListOptions)
func WithCommentsOrder(order CommentsOrder) CommentListOption

// Topic list options
type TopicListOption func(*topicListOptions)
func WithTopicsOrder(order TopicsOrder) TopicListOption
func WithTopicQuery(q string) TopicListOption

// Collection list options
type CollectionListOption func(*collectionListOptions)
func WithCollectionsOrder(order CollectionsOrder) CollectionListOption
func WithCollectionsFeatured(featured bool) CollectionListOption
func WithCollectionsUserID(userID string) CollectionListOption
func WithCollectionsPostID(postID string) CollectionListOption

// Search options
type SearchOption func(*searchOptions)
func WithSearchType(t SearchType) SearchOption    // default: all types
func WithSearchFirst(n int) SearchOption

// Collection creation params
type CreateCollectionParams struct {
    Name        string
    Tagline     string
    Description string
}

// Trend analysis options
type TrendOption func(*trendOptions)
func WithTrendDateRange(from, to time.Time) TrendOption // default: last 7 days
func WithTrendTopics(slugs []string) TrendOption        // filter by topics
func WithTrendTopN(n int) TrendOption                   // top N keywords, default 20
func WithTrendStopWords(words []string) TrendOption
```

---

## Trend Analysis (`trends.go`)

The `AnalyzeTrends` method is a high-level aggregation that paginates through posts
across the given date range and computes analytics in-memory. This is the core
feature for the GTM agent's product-market fit research.

```go
type TrendReport struct {
    DateRange       [2]time.Time       `json:"dateRange"`
    PostsAnalyzed   int                `json:"postsAnalyzed"`
    TotalVotes      int                `json:"totalVotes"`
    TotalComments   int                `json:"totalComments"`
    AvgVotes        float64            `json:"avgVotes"`
    AvgComments     float64            `json:"avgComments"`
    TopProducts     []PostSummary      `json:"topProducts"`     // by votes
    TopKeywords     []KeywordFreq      `json:"topKeywords"`     // from names + taglines
    TopTopics       []TopicFreq        `json:"topTopics"`       // most common topics
    RisingProducts  []PostSummary      `json:"risingProducts"`  // high vote velocity
    PeakLaunchDays  []DayOfWeek        `json:"peakLaunchDays"`  // best days to launch
    MakerActivity   []MakerSummary     `json:"makerActivity"`   // most prolific makers
}

type PostSummary struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Tagline   string  `json:"tagline"`
    Votes     int     `json:"votes"`
    Comments  int     `json:"comments"`
    DailyRank int     `json:"dailyRank"`
    URL       string  `json:"url"`
    Website   string  `json:"website"`
}

type KeywordFreq struct {
    Term  string `json:"term"`
    Count int    `json:"count"`
}

type TopicFreq struct {
    Slug  string `json:"slug"`
    Name  string `json:"name"`
    Count int    `json:"count"`
}

type DayOfWeek struct {
    Day   string `json:"day"`   // "Monday", "Tuesday", etc.
    Count int    `json:"count"`
}

type MakerSummary struct {
    UserID   string `json:"userId"`
    Username string `json:"username"`
    Name     string `json:"name"`
    Products int    `json:"products"`
    AvgVotes float64 `json:"avgVotes"`
}
```

---

## User Stories & Acceptance Criteria

### US-1 · Auth & Session Validation
**As a developer, I can construct a `Client` from a cookie struct so the session is validated and all subsequent calls are authenticated.**

- AC-1.1: `New(cookies)` sends a test GraphQL query (e.g., `ViewerQuery`) to validate the session. Returns `ErrUnauthorized` if the session is invalid.
- AC-1.2: `Cookies.Session` is validated as non-empty; returns `ErrInvalidAuth` if blank.
- AC-1.3: All cookies are attached to every request via the `Cookie` header.
- AC-1.4: Cloudflare clearance cookies are included; if expired, returns `ErrCFChallenge` with guidance to refresh.
- AC-1.5: `GetViewer()` returns the authenticated user after successful construction.

---

### US-2 · Get Homefeed (Today's Products)
**As a GTM agent, I can retrieve today's featured products so I can monitor what's getting traction.**

- AC-2.1: `GetHomefeed(ctx)` returns today's featured posts ranked by votes.
- AC-2.2: Each `Post` includes: ID, Slug, Name, Tagline, VotesCount, CommentsCount, DailyRank, Makers, Topics.
- AC-2.3: `Page.HasNext` + `WithAfter(cursor)` supports full pagination.
- AC-2.4: `WithFeatured(true)` filters to featured-only posts (default: featured only).

---

### US-3 · Get Posts by Date
**As a GTM agent, I can retrieve products launched on any specific date so I can study historical launch patterns.**

- AC-3.1: `GetPostsByDate(ctx, time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC))` returns posts from April 15, 2026.
- AC-3.2: `WithPostsOrder(PostsOrderVotes)` sorts by vote count descending.
- AC-3.3: Returns empty page (no error) when no posts exist for the given date.

---

### US-4 · Get Trending Posts
**As a GTM agent, I can see what's trending right now so I can identify emerging products with high vote velocity.**

- AC-4.1: `GetTrendingPosts(ctx)` returns posts with recent high engagement.
- AC-4.2: Results reflect time-weighted popularity, not just total votes.

---

### US-5 · Get New Posts
**As a GTM agent, I can browse the newest posts (chronological, no ranking) so I can find products early before they trend.**

- AC-5.1: `GetNewPosts(ctx)` returns posts in reverse chronological order.
- AC-5.2: Includes both featured and non-featured posts.

---

### US-6 · Get Single Post
**As a GTM agent, I can retrieve full details for any product so I can analyze it deeply.**

- AC-6.1: `GetPost(ctx, "chatgpt")` (by slug) returns the full `*Post` with all fields populated.
- AC-6.2: `GetPost(ctx, "12345")` (by ID) also works.
- AC-6.3: `Post.Makers`, `Post.Topics`, `Post.Media`, `Post.ProductLinks` are populated.
- AC-6.4: Returns `ErrNotFound` for nonexistent slugs/IDs.

---

### US-7 · Get Post Comments
**As a GTM agent, I can read all comments on a product so I can understand user feedback and sentiment.**

- AC-7.1: `GetPostComments(ctx, postID)` returns first page of comments.
- AC-7.2: Each `Comment` has: Body, User, VotesCount, CreatedAt, ParentID (for threading).
- AC-7.3: Replies are nested in `Comment.Replies` for threaded display.
- AC-7.4: `WithCommentsOrder(CommentsOrderOldest)` sorts chronologically.
- AC-7.5: Full pagination via `WithAfter(cursor)`.

---

### US-8 · Create Comment
**As a GTM agent, I can comment on a product so I can engage with makers and users.**

- AC-8.1: `CreateComment(ctx, postID, "Great product!")` returns the created `*Comment`.
- AC-8.2: `ReplyToComment(ctx, postID, parentID, "Thanks!")` creates a threaded reply.
- AC-8.3: Returns `ErrForbidden` if commenting is restricted.
- AC-8.4: Returns `ErrNotFound` for nonexistent postIDs.

---

### US-9 · Update / Delete Comment
**As a GTM agent, I can edit or delete my own comments.**

- AC-9.1: `UpdateComment(ctx, commentID, "Updated text")` returns the updated `*Comment`.
- AC-9.2: `DeleteComment(ctx, commentID)` returns `nil` on success.
- AC-9.3: Returns `ErrForbidden` when trying to edit/delete another user's comment.

---

### US-10 · Upvote / Remove Upvote
**As a GTM agent, I can upvote products so I can signal support for relevant launches.**

- AC-10.1: `Upvote(ctx, postID)` returns `nil` on success.
- AC-10.2: `RemoveUpvote(ctx, postID)` returns `nil` on success.
- AC-10.3: Upvoting an already-upvoted post returns `ErrAlreadyVoted`.
- AC-10.4: Removing upvote from a non-voted post returns `ErrNotVoted`.

---

### US-11 · Get Post Voters
**As a GTM agent, I can see who upvoted a product so I can identify engaged early adopters.**

- AC-11.1: `GetPostVoters(ctx, postID)` returns paginated list of `Vote` with `User` details.
- AC-11.2: `Vote.CreatedAt` is populated for each voter.

---

### US-12 · Get Post Reviews
**As a GTM agent, I can read product reviews so I can understand real user satisfaction.**

- AC-12.1: `GetPostReviews(ctx, postID)` returns reviews with Body, Rating, Sentiment, User.
- AC-12.2: `Review.Rating` is a float (e.g. 4.5 out of 5).

---

### US-13 · Create Review
**As a GTM agent, I can write a product review to share insights.**

- AC-13.1: `CreateReview(ctx, postID, "Excellent tool", 5.0)` returns the created `*Review`.
- AC-13.2: Rating must be between 1.0 and 5.0; returns `ErrInvalidParams` otherwise.

---

### US-14 · Browse Topics
**As a GTM agent, I can browse and search topics so I can identify product categories relevant to my market.**

- AC-14.1: `GetTopics(ctx)` returns paginated list of all topics.
- AC-14.2: `WithTopicQuery("artificial intelligence")` filters topics by name.
- AC-14.3: `WithTopicsOrder(TopicsOrderFollowers)` sorts by follower count.
- AC-14.4: Each `Topic` includes: Slug, Name, Description, PostsCount, FollowersCount.

---

### US-15 · Get Topic & Topic Posts
**As a GTM agent, I can drill into a topic to see what products are launching in that category.**

- AC-15.1: `GetTopic(ctx, "artificial-intelligence")` returns the full `*Topic`.
- AC-15.2: `GetTopicPosts(ctx, "artificial-intelligence")` returns posts tagged with that topic.
- AC-15.3: Supports `WithPostsOrder` and pagination.

---

### US-16 · Follow / Unfollow Topics
**As a GTM agent, I can follow topics to build my PH profile around my market focus.**

- AC-16.1: `FollowTopic(ctx, topicID)` returns `nil` on success.
- AC-16.2: `UnfollowTopic(ctx, topicID)` returns `nil` on success.

---

### US-17 · Browse Collections
**As a GTM agent, I can browse curated product collections so I can find thematic groupings of relevant products.**

- AC-17.1: `GetCollections(ctx)` returns paginated collections.
- AC-17.2: `WithCollectionsFeatured(true)` filters to featured collections.
- AC-17.3: `WithCollectionsUserID(id)` filters to collections by a specific user.

---

### US-18 · Manage Collections
**As a GTM agent, I can create and manage collections to organize products for my research.**

- AC-18.1: `CreateCollection(ctx, params)` returns the created `*Collection`.
- AC-18.2: `AddPostToCollection(ctx, collectionID, postID)` returns `nil` on success.
- AC-18.3: `RemovePostFromCollection` returns `nil` on success.
- AC-18.4: `FollowCollection` / `UnfollowCollection` work as expected.

---

### US-19 · User Profiles & Social Graph
**As a GTM agent, I can look up user profiles and their social connections so I can identify key makers, influencers, and early adopters.**

- AC-19.1: `GetUser(ctx, "rrhoover")` returns the user's full profile.
- AC-19.2: `GetUserPosts(ctx, userID)` returns products the user has made.
- AC-19.3: `GetUserVotedPosts(ctx, userID)` returns products the user has upvoted.
- AC-19.4: `GetUserFollowers(ctx, userID)` returns who follows the user.
- AC-19.5: `GetUserFollowing(ctx, userID)` returns who the user follows.

---

### US-20 · Follow / Unfollow Users
**As a GTM agent, I can follow users to build relationships and track their activity.**

- AC-20.1: `FollowUser(ctx, userID)` returns `nil` on success.
- AC-20.2: `UnfollowUser(ctx, userID)` returns `nil` on success.

---

### US-21 · Search
**As a GTM agent, I can search Product Hunt by keyword so I can find relevant products, users, and collections.**

- AC-21.1: `Search(ctx, "AI writing assistant")` returns results across posts, users, collections.
- AC-21.2: `WithSearchType(SearchPosts)` restricts to post results only.
- AC-21.3: Pagination is supported for each result type.

---

### US-22 · Get Viewer (Current User)
**As a developer, I can retrieve the currently authenticated user so I can verify identity and display context.**

- AC-22.1: `GetViewer(ctx)` returns the full `*User` for the authenticated session.
- AC-22.2: Returns `ErrUnauthorized` if the session has expired.

---

### US-23 · Analyze Trends
**As a GTM agent, I can run trend analysis across Product Hunt so I can identify market patterns, winning categories, and optimal launch timing.**

- AC-23.1: `AnalyzeTrends(ctx)` paginates through posts in the given date range (default: last 7 days).
- AC-23.2: `TrendReport.TopProducts` lists the highest-voted products in the period.
- AC-23.3: `TrendReport.TopKeywords` extracts unigram/bigram frequencies from product names + taglines, filtered for stop words.
- AC-23.4: `TrendReport.TopTopics` shows which categories had the most launches.
- AC-23.5: `TrendReport.RisingProducts` identifies products with unusually high vote velocity.
- AC-23.6: `TrendReport.PeakLaunchDays` shows which day of the week gets the most launches/votes.
- AC-23.7: `TrendReport.MakerActivity` identifies the most prolific makers by product count + average votes.
- AC-23.8: `WithTrendDateRange(from, to)` allows custom date ranges.
- AC-23.9: `WithTrendTopics([]string{"artificial-intelligence"})` filters to specific topics.
- AC-23.10: The function respects `ctx` cancellation; returns partial report with `ErrPartialResult`.

---

## Client Transport & Rate-Limiting

| Concern | Behaviour |
|---|---|
| **Request gap** | Min 1s between requests (leaky-bucket, configurable via `WithRateLimit`) |
| **Retry** | 3 attempts, 500ms exponential base; retries on 429 and 5xx only |
| **429 handling** | Honour `Retry-After` header; if absent, back off 60s |
| **CF challenge** | On 403 with Cloudflare challenge, return `ErrCFChallenge` — caller must refresh cookies |
| **Session expiry** | On 401/unauthorized GraphQL error, return `ErrUnauthorized` |
| **Proxy support** | `WithProxy("http://host:port")` wraps transport |
| **Context** | All methods accept `context.Context`; cancel propagates immediately |

---

## Sentinel Errors (`errors.go`)

```go
var (
    ErrInvalidAuth   = errors.New("producthunt: missing or empty session cookie")
    ErrUnauthorized  = errors.New("producthunt: authentication failed (session expired or invalid)")
    ErrCFChallenge   = errors.New("producthunt: Cloudflare challenge required — refresh cf_clearance cookie")
    ErrForbidden     = errors.New("producthunt: access denied")
    ErrNotFound      = errors.New("producthunt: resource not found")
    ErrRateLimited   = errors.New("producthunt: rate limited")
    ErrAlreadyVoted  = errors.New("producthunt: already upvoted this post")
    ErrNotVoted      = errors.New("producthunt: not voted on this post")
    ErrInvalidParams = errors.New("producthunt: invalid or missing required parameters")
    ErrPartialResult = errors.New("producthunt: context cancelled; partial result returned")
    ErrGraphQL       = errors.New("producthunt: GraphQL error")
)
```

---

## GTM Agent Integration Points

This client is designed for the go-to-market agent. Key workflows:

### 1. Product-Market Fit Research
```
AnalyzeTrends(last 30 days, topic="developer-tools")
→ identify top keywords, rising products, engagement patterns
→ compare against own product positioning
```

### 2. User Feedback Mining
```
GetTopicPosts(topic) → for each post → GetPostComments()
→ extract pain points, feature requests, sentiment
→ identify what users love/hate about competitors
```

### 3. Competitive Intelligence
```
Search("competitor-name") → GetPost(slug) → GetPostComments + GetPostReviews
→ understand competitor reception, user feedback, feature gaps
```

### 4. Community Engagement
```
GetHomefeed() → identify relevant launches → Upvote + CreateComment
→ build presence, connect with makers
FollowUser(maker) → track their future launches
```

### 5. Launch Timing Optimization
```
AnalyzeTrends(last 90 days)
→ TrendReport.PeakLaunchDays → identify best day/time to launch
→ TrendReport.TopTopics → identify hot categories to position in
```

---

## Out of Scope (v1)

- Product submission / launching (POST /submit flow)
- Ship / milestone updates
- Maker tools / analytics dashboard
- Product Hunt Discussions
- Notifications / inbox
- Product Hunt Ads
- OAuth token-based authentication (public API v2)
- WebSocket / real-time subscriptions

---

## Authentication (Verified)

The client supports **two authentication modes**, discovered during live testing:

### Mode 1: Developer Token (v2 API) — Recommended for reads

- Endpoint: `https://api.producthunt.com/v2/api/graphql`
- Auth: `Authorization: Bearer <developer_token>`
- No Cloudflare challenge — works with plain HTTP clients and curl.
- Get a token at: https://www.producthunt.com/v2/oauth/applications
- Supports all read operations. Write operations need a user-scoped OAuth token.

### Mode 2: Browser Cookies (Frontend API) — Required for writes

- Endpoint: `https://www.producthunt.com/frontend/graphql`
- Auth: Cookie-based (`_producthunt_session_production`, `csrf_token`)
- **Blocked by Cloudflare managed challenge** for plain HTTP clients (curl, Go net/http).
- Browser testing confirmed the frontend API is active — the site makes GraphQL calls
  to `/frontend/graphql` including `HeaderDesktopProductsNavigationQuery` and others.
- Cloudflare uses a newer "managed challenge" that solves via JS execution + TLS
  fingerprint. No `cf_clearance` cookie is issued — `__cf_bm` alone is not sufficient.
- To use from Go: requires TLS fingerprint spoofing (e.g. `utls` library) or a
  browser-based proxy. This is a future enhancement.

### Required Cookies (from browser session export)

| Cookie | Domain | Purpose |
|---|---|---|
| `_producthunt_session_production` | `.producthunt.com` | **Rails session** — primary auth |
| `csrf_token` | `www.producthunt.com` | **CSRF token** — needed for mutations |
| `__cf_bm` | `.producthunt.com` | Cloudflare bot management (30 min TTL) |
| `ajs_user_id` | `.producthunt.com` | User ID (`702087` for current session) |

### Cloudflare Findings (from curl testing 2026-04-20)

- `www.producthunt.com` responds with HTTP 403 + Cloudflare JS challenge to all
  non-browser clients, regardless of cookies carried.
- `api.producthunt.com` has **no** Cloudflare challenge — responds normally with
  proper auth.
- The v2 API uses the same GraphQL schema as the frontend API for all documented
  queries and mutations.

---

## Implementation Status

All 15 source files implemented (3,027 lines), compiles clean, zero external deps.

| File | Lines | Surface |
|---|---|---|
| `client.go` | 592 | Client, New(), dual-auth, Options, HTTP transport, rate limiter |
| `graphql.go` | 401 | Request/response types, Relay helpers, GQL→domain converters |
| `types.go` | 264 | All domain types, Cookies with DeveloperToken + cookie fields |
| `trends.go` | 311 | AnalyzeTrends — keywords, topics, rising products, timing |
| `users.go` | 291 | GetUser, posts, voted, followers, following, follow/unfollow |
| `collections.go` | 261 | CRUD + follow/unfollow + post management |
| `posts.go` | 237 | Homefeed, by-date, trending, new, launches, single post |
| `comments.go` | 187 | Get, create, reply, update, delete |
| `topics.go` | 156 | Browse, search, posts-by-topic, follow/unfollow |
| `reviews.go` | 112 | Get reviews, create review |
| `votes.go` | 90 | Get voters, upvote, remove upvote |
| `search.go` | 57 | Search posts (upgrade path for full-text) |
| `viewer.go` | 46 | Get authenticated user |
| `doc.go` | 34 | Package documentation |
| `errors.go` | 18 | 12 sentinel errors |
| `cmd/probe/main.go` | 86 | Integration test probe |

---

## Next Steps

1. **Get developer token** from https://www.producthunt.com/v2/oauth/applications
2. **Run probe**: `PH_DEV_TOKEN=xxx go run ./cmd/probe`
3. **Verify GraphQL field names** against actual API responses — fix any schema mismatches
4. **Add TLS fingerprint bypass** (utls) for frontend API access (writes)
5. **Integration tests** against live API
