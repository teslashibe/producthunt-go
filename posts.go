package producthunt

import (
	"context"
	"fmt"
	"time"
	"unicode"
)

// --- GraphQL query fragments ---

const postFields = `
  id
  slug
  name
  tagline
  description
  url
  website
  votesCount
  commentsCount
  reviewsCount
  reviewsRating
  dailyRank
  weeklyRank
  monthlyRank
  yearlyRank
  featuredAt
  createdAt
  scheduledAt
  isVoted
  isCollected
  thumbnail { type url videoUrl }
  media { type url videoUrl }
  productLinks { type url }
  topics(first: 10) {
    edges { node { id slug name description postsCount followersCount isFollowing url } }
  }
  makers { id username name headline profileImage url isMaker }
  user { id username name headline profileImage url isMaker }
`

const homefeedQuery = `
query HomefeedQuery($first: Int, $after: String, $featured: Boolean) {
  posts(first: $first, after: $after, featured: $featured, order: RANKING) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + postFields + `} }
  }
}
`

const postsByDateQuery = `
query PostsByDateQuery($postedAfter: DateTime, $postedBefore: DateTime, $featured: Boolean, $order: PostsOrder, $topic: String, $first: Int, $after: String) {
  posts(postedAfter: $postedAfter, postedBefore: $postedBefore, featured: $featured, order: $order, topic: $topic, first: $first, after: $after) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + postFields + `} }
  }
}
`

const singlePostBySlugQuery = `
query PostQuery($slug: String!) {
  post(slug: $slug) {` + postFields + `}
}
`

const singlePostByIDQuery = `
query PostQuery($id: ID!) {
  post(id: $id) {` + postFields + `}
}
`

// GetHomefeed returns today's featured/hot products.
func (c *Client) GetHomefeed(ctx context.Context, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)
	vars := map[string]interface{}{
		"first": o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}
	featured := true
	if o.featured != nil {
		featured = *o.featured
	}
	vars["featured"] = featured

	data, err := c.query(ctx, "HomefeedQuery", homefeedQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}
	return parsePostConnection(data)
}

// GetPostsByDate returns products launched on a specific date.
func (c *Client) GetPostsByDate(ctx context.Context, date time.Time, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	vars := map[string]interface{}{
		"postedAfter":  start.Format(time.RFC3339),
		"postedBefore": end.Format(time.RFC3339),
		"first":        o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}
	if o.order != "" {
		vars["order"] = string(o.order)
	}
	if o.featured != nil {
		vars["featured"] = *o.featured
	}
	if o.topic != "" {
		vars["topic"] = o.topic
	}

	data, err := c.query(ctx, "PostsByDateQuery", postsByDateQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}
	return parsePostConnection(data)
}

// GetTrendingPosts returns posts that are currently trending (high recent engagement).
func (c *Client) GetTrendingPosts(ctx context.Context, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)
	vars := map[string]interface{}{
		"first":   o.first,
		"order":   "RANKING",
		"featured": true,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "PostsByDateQuery", postsByDateQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}
	return parsePostConnection(data)
}

// GetNewPosts returns the newest posts in reverse chronological order.
func (c *Client) GetNewPosts(ctx context.Context, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)
	vars := map[string]interface{}{
		"first": o.first,
		"order": "NEWEST",
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "PostsByDateQuery", postsByDateQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}
	return parsePostConnection(data)
}

// GetLaunches returns products launched on a specific date.
func (c *Client) GetLaunches(ctx context.Context, date time.Time, opts ...PostListOption) (Page[Post], error) {
	return c.GetPostsByDate(ctx, date, opts...)
}

// GetPost returns a single post by slug or ID.
// If the input looks numeric it's treated as an ID, otherwise a slug.
func (c *Client) GetPost(ctx context.Context, slugOrID string) (*Post, error) {
	if slugOrID == "" {
		return nil, fmt.Errorf("%w: slugOrID must be non-empty", ErrInvalidParams)
	}

	var q string
	var vars map[string]interface{}

	if isNumeric(slugOrID) {
		q = singlePostByIDQuery
		vars = map[string]interface{}{"id": slugOrID}
	} else {
		q = singlePostBySlugQuery
		vars = map[string]interface{}{"slug": slugOrID}
	}

	data, err := c.query(ctx, "PostQuery", q, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Post *gqlPost `json:"post"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.Post == nil {
		return nil, ErrNotFound
	}

	p := resp.Post.toPost()
	return &p, nil
}

// --- helpers ---

func parsePostConnection(data []byte) (Page[Post], error) {
	var resp struct {
		Posts relayConnection[gqlPost] `json:"posts"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Post]{}, err
	}

	page := Page[Post]{
		TotalCount: resp.Posts.TotalCount,
		HasNext:    resp.Posts.PageInfo.HasNextPage,
		EndCursor:  resp.Posts.PageInfo.EndCursor,
	}
	page.Items = make([]Post, 0, len(resp.Posts.Edges))
	for _, e := range resp.Posts.Edges {
		page.Items = append(page.Items, e.Node.toPost())
	}
	return page, nil
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}
