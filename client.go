package producthunt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	frontendEndpoint = "https://www.producthunt.com/frontend/graphql"
	v2Endpoint       = "https://api.producthunt.com/v2/api/graphql"
	phOrigin         = "https://www.producthunt.com"
	phReferer        = "https://www.producthunt.com/"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	defaultMinGap     = 1 * time.Second
	defaultMaxRetries = 3
	defaultRetryBase  = 500 * time.Millisecond
)

// Client is a Product Hunt API client. It is safe for concurrent use.
type Client struct {
	cookies    Cookies
	endpoint   string // resolved API endpoint (v2 or frontend)
	useBearer  bool   // true when using developer token + v2 API
	httpClient *http.Client
	userAgent  string
	minGap     time.Duration
	maxRetries int
	retryBase  time.Duration

	mu        sync.Mutex
	lastReqAt time.Time
}

// New constructs a Client and validates the session by calling GetViewer.
//
// If Cookies.DeveloperToken is set, the client uses the public v2 API
// (api.producthunt.com) with Bearer token auth — no Cloudflare issues.
//
// If only Cookies.Session is set, the client uses the internal frontend API
// (www.producthunt.com/frontend/graphql) with cookie auth. This requires a
// cf_clearance cookie or TLS fingerprint spoofing to bypass Cloudflare.
//
// Returns ErrInvalidAuth if neither DeveloperToken nor Session is provided.
func New(cookies Cookies, opts ...Option) (*Client, error) {
	hasDev := strings.TrimSpace(cookies.DeveloperToken) != ""
	hasSess := strings.TrimSpace(cookies.Session) != ""

	if !hasDev && !hasSess {
		return nil, fmt.Errorf("%w: either DeveloperToken or Session cookie must be non-empty", ErrInvalidAuth)
	}

	endpoint := frontendEndpoint
	useBearer := false
	if hasDev {
		endpoint = v2Endpoint
		useBearer = true
	}

	c := &Client{
		cookies:    cookies,
		endpoint:   endpoint,
		useBearer:  useBearer,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  defaultUserAgent,
		minGap:     defaultMinGap,
		maxRetries: defaultMaxRetries,
		retryBase:  defaultRetryBase,
	}
	for _, o := range opts {
		o(c)
	}

	if _, err := c.GetViewer(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

// Option configures a Client.
type Option func(*Client)

// WithUserAgent overrides the default Chrome User-Agent string.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithHTTPClient replaces the default http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithProxy routes all HTTP traffic through the given proxy URL.
func WithProxy(proxyURL string) Option {
	return func(c *Client) {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(parsed)
		c.httpClient = &http.Client{
			Timeout:   c.httpClient.Timeout,
			Transport: transport,
		}
	}
}

// WithRateLimit sets the minimum time between consecutive requests. Default: 1s.
func WithRateLimit(d time.Duration) Option {
	return func(c *Client) { c.minGap = d }
}

// WithRetry configures retry behaviour. Default: 3 attempts, 500ms exponential base.
func WithRetry(maxAttempts int, base time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxAttempts
		c.retryBase = base
	}
}

// --- GraphQL transport ---

// query executes an authenticated GraphQL query/mutation against the frontend API.
func (c *Client) query(ctx context.Context, operationName, gqlQuery string, variables interface{}) (json.RawMessage, error) {
	attempts := c.maxRetries
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			wait := c.retryBase * time.Duration(math.Pow(2, float64(i-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		data, err := c.doQuery(ctx, operationName, gqlQuery, variables)
		if err == nil {
			return data, nil
		}
		if isNonRetriable(err) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

// doQuery performs a single JSON POST to the frontend GraphQL endpoint.
func (c *Client) doQuery(ctx context.Context, operationName, gqlQuery string, variables interface{}) (json.RawMessage, error) {
	c.waitForGap(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	body, err := json.Marshal(gqlRequest{
		OperationName: operationName,
		Variables:     variables,
		Query:         gqlQuery,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: marshalling request: %v", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: building request: %v", ErrRequestFailed, err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		// handled below
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case resp.StatusCode == http.StatusForbidden:
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "text/html") {
			return nil, ErrCFChallenge
		}
		return nil, ErrForbidden
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), 60*time.Second)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		return nil, ErrRateLimited
	case resp.StatusCode >= 500:
		return nil, fmt.Errorf("%w: HTTP %d", ErrRequestFailed, resp.StatusCode)
	default:
		return nil, fmt.Errorf("%w: unexpected HTTP %d", ErrRequestFailed, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, err)
	}

	var envelope gqlResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("%w: decoding envelope: %v (body: %s)",
			ErrRequestFailed, err, truncate(string(respBody), 200))
	}

	if err := envelope.err(); err != nil {
		return nil, err
	}

	return envelope.Data, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", c.userAgent)

	if c.useBearer {
		req.Header.Set("Authorization", "Bearer "+c.cookies.DeveloperToken)
	} else {
		req.Header.Set("Origin", phOrigin)
		req.Header.Set("Referer", phReferer)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Cookie", c.cookieHeader())
		if c.cookies.CSRFToken != "" {
			req.Header.Set("X-CSRF-Token", c.cookies.CSRFToken)
		}
	}
}

func (c *Client) cookieHeader() string {
	var b strings.Builder
	add := func(name, val string) {
		if val == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteString(name)
		b.WriteByte('=')
		b.WriteString(val)
	}
	add("_producthunt_session_production", c.cookies.Session)
	add("cf_clearance", c.cookies.CFClearance)
	add("__cf_bm", c.cookies.CFBM)
	add("csrf_token", c.cookies.CSRFToken)
	add("_ph_id", c.cookies.PHID)
	return b.String()
}

func (c *Client) waitForGap(ctx context.Context) {
	c.mu.Lock()
	since := time.Since(c.lastReqAt)
	if since < c.minGap {
		wait := c.minGap - since
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		c.mu.Lock()
	}
	c.lastReqAt = time.Now()
	c.mu.Unlock()
}

func isNonRetriable(err error) bool {
	switch {
	case strings.Contains(err.Error(), ErrInvalidAuth.Error()),
		strings.Contains(err.Error(), ErrUnauthorized.Error()),
		strings.Contains(err.Error(), ErrCFChallenge.Error()),
		strings.Contains(err.Error(), ErrForbidden.Error()),
		strings.Contains(err.Error(), ErrNotFound.Error()),
		strings.Contains(err.Error(), ErrInvalidParams.Error()),
		strings.Contains(err.Error(), ErrAlreadyVoted.Error()),
		strings.Contains(err.Error(), ErrNotVoted.Error()):
		return true
	}
	return false
}

func parseRetryAfter(val string, fallback time.Duration) time.Duration {
	if val == "" {
		return fallback
	}
	if secs, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(val); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return fallback
}

// --- Pagination option types ---

// PageOption configures pagination for list queries.
type PageOption func(*pageOptions)

type pageOptions struct {
	first int
	after string
}

func defaultPageOptions() pageOptions {
	return pageOptions{first: 20}
}

// WithFirst sets the number of items per page. Default: 20, max: 50.
func WithFirst(n int) PageOption {
	return func(o *pageOptions) {
		if n > 50 {
			n = 50
		}
		o.first = n
	}
}

// WithAfter sets the cursor for the next page.
func WithAfter(cursor string) PageOption {
	return func(o *pageOptions) { o.after = cursor }
}

// PostListOption configures post list queries.
type PostListOption func(*postListOptions)

type postListOptions struct {
	pageOptions
	order        PostsOrder
	featured     *bool
	postedAfter  *time.Time
	postedBefore *time.Time
	topic        string
}

func defaultPostListOptions() postListOptions {
	return postListOptions{pageOptions: defaultPageOptions()}
}

func applyPostListOptions(opts []PostListOption) postListOptions {
	o := defaultPostListOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithPostListFirst sets items per page for post list queries.
func WithPostListFirst(n int) PostListOption {
	return func(o *postListOptions) {
		if n > 50 {
			n = 50
		}
		o.first = n
	}
}

// WithPostListAfter sets the pagination cursor for post list queries.
func WithPostListAfter(cursor string) PostListOption {
	return func(o *postListOptions) { o.after = cursor }
}

// WithPostsOrder sets the sort order for post queries.
func WithPostsOrder(order PostsOrder) PostListOption {
	return func(o *postListOptions) { o.order = order }
}

// WithFeatured filters to featured or non-featured posts.
func WithFeatured(featured bool) PostListOption {
	return func(o *postListOptions) { o.featured = &featured }
}

// WithPostedAfter filters to posts created after the given time.
func WithPostedAfter(t time.Time) PostListOption {
	return func(o *postListOptions) { o.postedAfter = &t }
}

// WithPostedBefore filters to posts created before the given time.
func WithPostedBefore(t time.Time) PostListOption {
	return func(o *postListOptions) { o.postedBefore = &t }
}

// WithPostsTopic filters posts by topic slug.
func WithPostsTopic(slug string) PostListOption {
	return func(o *postListOptions) { o.topic = slug }
}

// CommentListOption configures comment list queries.
type CommentListOption func(*commentListOptions)

type commentListOptions struct {
	pageOptions
	order CommentsOrder
}

func defaultCommentListOptions() commentListOptions {
	return commentListOptions{pageOptions: defaultPageOptions(), order: CommentsOrderNewest}
}

func applyCommentListOptions(opts []CommentListOption) commentListOptions {
	o := defaultCommentListOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithCommentsOrder sets the sort order for comment queries.
func WithCommentsOrder(order CommentsOrder) CommentListOption {
	return func(o *commentListOptions) { o.order = order }
}

// TopicListOption configures topic list queries.
type TopicListOption func(*topicListOptions)

type topicListOptions struct {
	pageOptions
	order TopicsOrder
	query string
}

func defaultTopicListOptions() topicListOptions {
	return topicListOptions{pageOptions: defaultPageOptions()}
}

func applyTopicListOptions(opts []TopicListOption) topicListOptions {
	o := defaultTopicListOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithTopicsOrder sets the sort order for topic queries.
func WithTopicsOrder(order TopicsOrder) TopicListOption {
	return func(o *topicListOptions) { o.order = order }
}

// WithTopicQuery filters topics by name or alias.
func WithTopicQuery(q string) TopicListOption {
	return func(o *topicListOptions) { o.query = q }
}

// CollectionListOption configures collection list queries.
type CollectionListOption func(*collectionListOptions)

type collectionListOptions struct {
	pageOptions
	order    CollectionsOrder
	featured *bool
	userID   string
	postID   string
}

func defaultCollectionListOptions() collectionListOptions {
	return collectionListOptions{pageOptions: defaultPageOptions()}
}

func applyCollectionListOptions(opts []CollectionListOption) collectionListOptions {
	o := defaultCollectionListOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithCollectionsOrder sets the sort order for collection queries.
func WithCollectionsOrder(order CollectionsOrder) CollectionListOption {
	return func(o *collectionListOptions) { o.order = order }
}

// WithCollectionsFeatured filters to featured collections.
func WithCollectionsFeatured(featured bool) CollectionListOption {
	return func(o *collectionListOptions) { o.featured = &featured }
}

// WithCollectionsUserID filters collections by creator.
func WithCollectionsUserID(userID string) CollectionListOption {
	return func(o *collectionListOptions) { o.userID = userID }
}

// WithCollectionsPostID filters collections containing the given post.
func WithCollectionsPostID(postID string) CollectionListOption {
	return func(o *collectionListOptions) { o.postID = postID }
}

// SearchOption configures search queries.
type SearchOption func(*searchOptions)

type searchOptions struct {
	searchType SearchType
	first      int
	after      string
}

func defaultSearchOptions() searchOptions {
	return searchOptions{first: 20}
}

func applySearchOptions(opts []SearchOption) searchOptions {
	o := defaultSearchOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithSearchType restricts search results to a specific type.
func WithSearchType(t SearchType) SearchOption {
	return func(o *searchOptions) { o.searchType = t }
}

// WithSearchFirst sets items per page for search. Default: 20.
func WithSearchFirst(n int) SearchOption {
	return func(o *searchOptions) { o.first = n }
}

// WithSearchAfter sets the pagination cursor for search.
func WithSearchAfter(cursor string) SearchOption {
	return func(o *searchOptions) { o.after = cursor }
}

// TrendOption configures AnalyzeTrends.
type TrendOption func(*trendOptions)

type trendOptions struct {
	from      time.Time
	to        time.Time
	topics    []string
	topN      int
	stopWords []string
}

func defaultTrendOptions() trendOptions {
	now := time.Now().UTC()
	return trendOptions{
		from: now.AddDate(0, 0, -7),
		to:   now,
		topN: 20,
	}
}

func applyTrendOptions(opts []TrendOption) trendOptions {
	o := defaultTrendOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithTrendDateRange sets the date range for trend analysis.
func WithTrendDateRange(from, to time.Time) TrendOption {
	return func(o *trendOptions) { o.from = from; o.to = to }
}

// WithTrendTopics filters trend analysis to specific topic slugs.
func WithTrendTopics(slugs []string) TrendOption {
	return func(o *trendOptions) { o.topics = slugs }
}

// WithTrendTopN sets the number of top items returned. Default: 20.
func WithTrendTopN(n int) TrendOption {
	return func(o *trendOptions) { o.topN = n }
}

// WithTrendStopWords adds domain-specific stop words to the keyword filter.
func WithTrendStopWords(words []string) TrendOption {
	return func(o *trendOptions) { o.stopWords = append(o.stopWords, words...) }
}

func applyPageOptions(opts []PageOption) pageOptions {
	o := defaultPageOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}
