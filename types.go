package producthunt

import "time"

// Credentials holds Product Hunt authentication credentials.
//
// Three authentication modes are supported (in priority order):
//
// 1. Developer token (BYOK): Set DeveloperToken to a token from
//    https://www.producthunt.com/v2/oauth/applications — never expires,
//    full user context, no Cloudflare issues.
//
// 2. Client credentials: Set ClientID + ClientSecret from the same page.
//    The client calls the OAuth token endpoint on construction to obtain
//    an access token automatically. Public scope only (no user context).
//
// 3. Browser cookies: Set Session (and CSRFToken) from a browser export.
//    Used for the frontend API (www.producthunt.com/frontend/graphql).
//    Requires Cloudflare bypass (cf_clearance or TLS fingerprint spoofing).
//
// If DeveloperToken is set it takes priority over ClientID/ClientSecret,
// which in turn takes priority over browser cookies.
type Credentials struct {
	DeveloperToken string // BYOK: OAuth2 developer token (v2 API, user scope, never expires)
	ClientID       string // OAuth2 client ID (for auto-provisioned client_credentials flow)
	ClientSecret   string // OAuth2 client secret
	Session        string // _producthunt_session_production (browser cookie)
	CFClearance    string // cf_clearance (Cloudflare challenge clearance)
	CFBM           string // __cf_bm (Cloudflare bot management)
	CSRFToken      string // csrf_token (Rails CSRF — needed for mutations via frontend API)
}

// Cookies is an alias for Credentials for backward compatibility.
type Cookies = Credentials

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
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Name           string    `json:"name"`
	Headline       string    `json:"headline,omitempty"`
	ProfileImage   string    `json:"profileImage,omitempty"`
	CoverImage     string    `json:"coverImage,omitempty"`
	TwitterUsername string   `json:"twitterUsername,omitempty"`
	WebsiteURL     string    `json:"websiteUrl,omitempty"`
	URL            string    `json:"url"`
	FollowersCount int       `json:"followersCount"`
	FollowingCount int       `json:"followingCount"`
	IsMaker        bool      `json:"isMaker"`
	IsFollowing    bool      `json:"isFollowing"`
	IsViewer       bool      `json:"isViewer"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
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
	ID             string    `json:"id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Image          string    `json:"image,omitempty"`
	PostsCount     int       `json:"postsCount"`
	FollowersCount int       `json:"followersCount"`
	IsFollowing    bool      `json:"isFollowing"`
	URL            string    `json:"url"`
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
	Sentiment string    `json:"sentiment,omitempty"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
}

// Media is an image or video associated with a post.
type Media struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	VideoURL string `json:"videoUrl,omitempty"`
}

// ProductLink is an additional link on a post (App Store, Play Store, etc.).
type ProductLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

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

// CreateCollectionParams holds parameters for creating a collection.
type CreateCollectionParams struct {
	Name        string
	Tagline     string
	Description string
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
	CollectionsOrderFeatured  CollectionsOrder = "FEATURED_AT"
	CollectionsOrderFollowers CollectionsOrder = "FOLLOWERS_COUNT"
	CollectionsOrderNewest    CollectionsOrder = "NEWEST"
)

// TrendReport is the output of AnalyzeTrends.
type TrendReport struct {
	DateRange      [2]time.Time   `json:"dateRange"`
	PostsAnalyzed  int            `json:"postsAnalyzed"`
	TotalVotes     int            `json:"totalVotes"`
	TotalComments  int            `json:"totalComments"`
	AvgVotes       float64        `json:"avgVotes"`
	AvgComments    float64        `json:"avgComments"`
	TopProducts    []PostSummary  `json:"topProducts"`
	TopKeywords    []KeywordFreq  `json:"topKeywords"`
	TopTopics      []TopicFreq    `json:"topTopics"`
	RisingProducts []PostSummary  `json:"risingProducts"`
	PeakLaunchDays []DayOfWeek    `json:"peakLaunchDays"`
	MakerActivity  []MakerSummary `json:"makerActivity"`
}

// PostSummary is a lightweight post representation for trend reports.
type PostSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Tagline   string `json:"tagline"`
	Votes     int    `json:"votes"`
	Comments  int    `json:"comments"`
	DailyRank int    `json:"dailyRank"`
	URL       string `json:"url"`
	Website   string `json:"website"`
}

// KeywordFreq is a keyword and its occurrence count.
type KeywordFreq struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

// TopicFreq is a topic and its post count in a trend period.
type TopicFreq struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DayOfWeek is a day-of-week with a post count.
type DayOfWeek struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

// MakerSummary is a maker's activity summary in a trend period.
type MakerSummary struct {
	UserID   string  `json:"userId"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Products int     `json:"products"`
	AvgVotes float64 `json:"avgVotes"`
}

// RateLimitState captures rate-limit information from the most recently observed
// response headers. All fields are zero-valued until a response with rate-limit
// headers is received.
type RateLimitState struct {
	Limit      int           `json:"limit"`       // max requests per window (0 = not reported)
	Remaining  int           `json:"remaining"`   // requests left in the current window
	Reset      time.Time     `json:"reset"`       // when the window resets (UTC)
	RetryAfter time.Duration `json:"retry_after"` // set to Retry-After duration after a 429
}

// IsLimited reports whether the current state indicates requests are blocked.
func (r RateLimitState) IsLimited() bool {
	if !r.Reset.IsZero() && r.Remaining == 0 && time.Now().Before(r.Reset) {
		return true
	}
	return r.RetryAfter > 0
}

// ResetIn returns how long until the rate-limit window resets.
// Returns 0 if Reset is in the past or not set.
func (r RateLimitState) ResetIn() time.Duration {
	if r.Reset.IsZero() {
		return 0
	}
	if d := time.Until(r.Reset); d > 0 {
		return d
	}
	return 0
}
