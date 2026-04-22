package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetHomefeedInput is the typed input for producthunt_get_homefeed.
type GetHomefeedInput struct {
	Featured *bool  `json:"featured,omitempty" jsonschema:"description=filter to featured posts (default true; pass false to include non-featured)"`
	First    int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After    string `json:"after,omitempty" jsonschema:"description=pagination cursor (endCursor from a previous page)"`
}

func getHomefeed(ctx context.Context, c *producthunt.Client, in GetHomefeedInput) (any, error) {
	opts := []producthunt.PostListOption{}
	if in.First > 0 {
		opts = append(opts, producthunt.WithPostListFirst(in.First))
	}
	if in.After != "" {
		opts = append(opts, producthunt.WithPostListAfter(in.After))
	}
	if in.Featured != nil {
		opts = append(opts, producthunt.WithFeatured(*in.Featured))
	}
	page, err := c.GetHomefeed(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetPostsByDateInput is the typed input for producthunt_get_posts_by_date.
type GetPostsByDateInput struct {
	Date     string `json:"date" jsonschema:"description=launch date in YYYY-MM-DD (UTC) or RFC3339,required"`
	Order    string `json:"order,omitempty" jsonschema:"description=sort order; allowed: RANKING,NEWEST,VOTES"`
	Featured *bool  `json:"featured,omitempty" jsonschema:"description=filter to featured posts only when set"`
	Topic    string `json:"topic,omitempty" jsonschema:"description=topic slug filter (e.g. 'artificial-intelligence')"`
	First    int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After    string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getPostsByDate(ctx context.Context, c *producthunt.Client, in GetPostsByDateInput) (any, error) {
	date, err := parseDate(in.Date)
	if err != nil {
		return nil, err
	}
	opts := buildPostListOpts(in.First, in.After, in.Order, in.Featured, in.Topic)
	page, err := c.GetPostsByDate(ctx, date, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetTrendingPostsInput is the typed input for producthunt_get_trending_posts.
type GetTrendingPostsInput struct {
	First int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getTrendingPosts(ctx context.Context, c *producthunt.Client, in GetTrendingPostsInput) (any, error) {
	opts := []producthunt.PostListOption{}
	if in.First > 0 {
		opts = append(opts, producthunt.WithPostListFirst(in.First))
	}
	if in.After != "" {
		opts = append(opts, producthunt.WithPostListAfter(in.After))
	}
	page, err := c.GetTrendingPosts(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetNewPostsInput is the typed input for producthunt_get_new_posts.
type GetNewPostsInput struct {
	First int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getNewPosts(ctx context.Context, c *producthunt.Client, in GetNewPostsInput) (any, error) {
	opts := []producthunt.PostListOption{}
	if in.First > 0 {
		opts = append(opts, producthunt.WithPostListFirst(in.First))
	}
	if in.After != "" {
		opts = append(opts, producthunt.WithPostListAfter(in.After))
	}
	page, err := c.GetNewPosts(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetLaunchesInput is the typed input for producthunt_get_launches.
type GetLaunchesInput struct {
	Date     string `json:"date" jsonschema:"description=launch date in YYYY-MM-DD (UTC) or RFC3339,required"`
	Order    string `json:"order,omitempty" jsonschema:"description=sort order; allowed: RANKING,NEWEST,VOTES"`
	Featured *bool  `json:"featured,omitempty" jsonschema:"description=filter to featured posts only when set"`
	Topic    string `json:"topic,omitempty" jsonschema:"description=topic slug filter"`
	First    int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After    string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getLaunches(ctx context.Context, c *producthunt.Client, in GetLaunchesInput) (any, error) {
	date, err := parseDate(in.Date)
	if err != nil {
		return nil, err
	}
	opts := buildPostListOpts(in.First, in.After, in.Order, in.Featured, in.Topic)
	page, err := c.GetLaunches(ctx, date, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetPostInput is the typed input for producthunt_get_post.
type GetPostInput struct {
	SlugOrID string `json:"slug_or_id" jsonschema:"description=post slug (e.g. 'cursor-2-0') or numeric ID,required"`
}

func getPost(ctx context.Context, c *producthunt.Client, in GetPostInput) (any, error) {
	return c.GetPost(ctx, in.SlugOrID)
}

func buildPostListOpts(first int, after, order string, featured *bool, topic string) []producthunt.PostListOption {
	var opts []producthunt.PostListOption
	if first > 0 {
		opts = append(opts, producthunt.WithPostListFirst(first))
	}
	if after != "" {
		opts = append(opts, producthunt.WithPostListAfter(after))
	}
	if order != "" {
		opts = append(opts, producthunt.WithPostsOrder(producthunt.PostsOrder(order)))
	}
	if featured != nil {
		opts = append(opts, producthunt.WithFeatured(*featured))
	}
	if topic != "" {
		opts = append(opts, producthunt.WithPostsTopic(topic))
	}
	return opts
}

var postsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetHomefeedInput](
		"producthunt_get_homefeed",
		"Fetch today's Product Hunt home feed (featured launches by ranking)",
		"GetHomefeed",
		getHomefeed,
	),
	mcptool.Define[*producthunt.Client, GetPostsByDateInput](
		"producthunt_get_posts_by_date",
		"Fetch Product Hunt posts launched on a specific date with order/featured/topic filters",
		"GetPostsByDate",
		getPostsByDate,
	),
	mcptool.Define[*producthunt.Client, GetTrendingPostsInput](
		"producthunt_get_trending_posts",
		"Fetch posts currently trending on Product Hunt (high recent engagement)",
		"GetTrendingPosts",
		getTrendingPosts,
	),
	mcptool.Define[*producthunt.Client, GetNewPostsInput](
		"producthunt_get_new_posts",
		"Fetch the newest Product Hunt posts in reverse chronological order",
		"GetNewPosts",
		getNewPosts,
	),
	mcptool.Define[*producthunt.Client, GetLaunchesInput](
		"producthunt_get_launches",
		"Fetch products launched on a specific date (alias of get_posts_by_date)",
		"GetLaunches",
		getLaunches,
	),
	mcptool.Define[*producthunt.Client, GetPostInput](
		"producthunt_get_post",
		"Fetch a single Product Hunt post by slug or numeric ID",
		"GetPost",
		getPost,
	),
}
