package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetTopicsInput is the typed input for producthunt_get_topics.
//
// Note: the underlying client does not expose page-size or cursor setters
// for topic listings, so results are capped at the default of 20 per call.
type GetTopicsInput struct {
	Query string `json:"query,omitempty" jsonschema:"description=filter topics by name or alias substring"`
	Order string `json:"order,omitempty" jsonschema:"description=sort order; allowed: NEWEST,FOLLOWERS_COUNT"`
}

func getTopics(ctx context.Context, c *producthunt.Client, in GetTopicsInput) (any, error) {
	var opts []producthunt.TopicListOption
	if in.Order != "" {
		opts = append(opts, producthunt.WithTopicsOrder(producthunt.TopicsOrder(in.Order)))
	}
	if in.Query != "" {
		opts = append(opts, producthunt.WithTopicQuery(in.Query))
	}
	page, err := c.GetTopics(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, 0), nil
}

// GetTopicInput is the typed input for producthunt_get_topic.
type GetTopicInput struct {
	Slug string `json:"slug" jsonschema:"description=Product Hunt topic slug (e.g. 'artificial-intelligence'),required"`
}

func getTopic(ctx context.Context, c *producthunt.Client, in GetTopicInput) (any, error) {
	return c.GetTopic(ctx, in.Slug)
}

// GetTopicPostsInput is the typed input for producthunt_get_topic_posts.
type GetTopicPostsInput struct {
	TopicSlug string `json:"topic_slug" jsonschema:"description=Product Hunt topic slug,required"`
	Order     string `json:"order,omitempty" jsonschema:"description=sort order; allowed: RANKING,NEWEST,VOTES"`
	First     int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After     string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getTopicPosts(ctx context.Context, c *producthunt.Client, in GetTopicPostsInput) (any, error) {
	opts := buildPostListOpts(in.First, in.After, in.Order, nil, "")
	page, err := c.GetTopicPosts(ctx, in.TopicSlug, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// FollowTopicInput is the typed input for producthunt_follow_topic.
type FollowTopicInput struct {
	TopicID string `json:"topic_id" jsonschema:"description=Product Hunt topic ID to follow,required"`
}

func followTopic(ctx context.Context, c *producthunt.Client, in FollowTopicInput) (any, error) {
	if err := c.FollowTopic(ctx, in.TopicID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "topic_id": in.TopicID}, nil
}

// UnfollowTopicInput is the typed input for producthunt_unfollow_topic.
type UnfollowTopicInput struct {
	TopicID string `json:"topic_id" jsonschema:"description=Product Hunt topic ID to unfollow,required"`
}

func unfollowTopic(ctx context.Context, c *producthunt.Client, in UnfollowTopicInput) (any, error) {
	if err := c.UnfollowTopic(ctx, in.TopicID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "topic_id": in.TopicID}, nil
}

var topicsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetTopicsInput](
		"producthunt_get_topics",
		"List Product Hunt topics with optional name filter and sort order",
		"GetTopics",
		getTopics,
	),
	mcptool.Define[*producthunt.Client, GetTopicInput](
		"producthunt_get_topic",
		"Fetch a single Product Hunt topic by slug",
		"GetTopic",
		getTopic,
	),
	mcptool.Define[*producthunt.Client, GetTopicPostsInput](
		"producthunt_get_topic_posts",
		"List Product Hunt posts tagged with a specific topic slug",
		"GetTopicPosts",
		getTopicPosts,
	),
	mcptool.Define[*producthunt.Client, FollowTopicInput](
		"producthunt_follow_topic",
		"Follow a Product Hunt topic as the authenticated viewer",
		"FollowTopic",
		followTopic,
	),
	mcptool.Define[*producthunt.Client, UnfollowTopicInput](
		"producthunt_unfollow_topic",
		"Unfollow a Product Hunt topic the authenticated viewer follows",
		"UnfollowTopic",
		unfollowTopic,
	),
}
