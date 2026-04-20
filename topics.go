package producthunt

import (
	"context"
	"fmt"
)

const topicFields = `
  id slug name description image postsCount followersCount isFollowing url createdAt
`

const topicsListQuery = `
query TopicsQuery($query: String, $order: TopicsOrder, $first: Int, $after: String) {
  topics(query: $query, order: $order, first: $first, after: $after) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + topicFields + `} }
  }
}
`

const singleTopicQuery = `
query TopicQuery($slug: String!) {
  topic(slug: $slug) {` + topicFields + `}
}
`

const topicPostsQuery = `
query TopicPostsQuery($topicSlug: String!, $order: PostsOrder, $first: Int, $after: String) {
  posts(topic: $topicSlug, order: $order, first: $first, after: $after) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + postFields + `} }
  }
}
`

const topicFollowMutation = `
mutation TopicFollowMutation($topicId: ID!) {
  topicFollow(input: { topicId: $topicId }) {
    node { id isFollowing }
  }
}
`

const topicUnfollowMutation = `
mutation TopicUnfollowMutation($topicId: ID!) {
  topicUnfollow(input: { topicId: $topicId }) {
    node { id isFollowing }
  }
}
`

// GetTopics returns a paginated list of topics.
func (c *Client) GetTopics(ctx context.Context, opts ...TopicListOption) (Page[Topic], error) {
	o := applyTopicListOptions(opts)
	vars := map[string]interface{}{
		"first": o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}
	if o.order != "" {
		vars["order"] = string(o.order)
	}
	if o.query != "" {
		vars["query"] = o.query
	}

	data, err := c.query(ctx, "TopicsQuery", topicsListQuery, vars)
	if err != nil {
		return Page[Topic]{}, err
	}

	var resp struct {
		Topics relayConnection[gqlTopic] `json:"topics"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Topic]{}, err
	}

	page := Page[Topic]{
		TotalCount: resp.Topics.TotalCount,
		HasNext:    resp.Topics.PageInfo.HasNextPage,
		EndCursor:  resp.Topics.PageInfo.EndCursor,
	}
	page.Items = make([]Topic, 0, len(resp.Topics.Edges))
	for _, e := range resp.Topics.Edges {
		page.Items = append(page.Items, e.Node.toTopic())
	}
	return page, nil
}

// GetTopic returns a single topic by slug.
func (c *Client) GetTopic(ctx context.Context, slug string) (*Topic, error) {
	if slug == "" {
		return nil, fmt.Errorf("%w: topic slug must be non-empty", ErrInvalidParams)
	}

	data, err := c.query(ctx, "TopicQuery", singleTopicQuery, map[string]interface{}{
		"slug": slug,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Topic *gqlTopic `json:"topic"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.Topic == nil {
		return nil, ErrNotFound
	}

	t := resp.Topic.toTopic()
	return &t, nil
}

// GetTopicPosts returns posts tagged with a specific topic.
func (c *Client) GetTopicPosts(ctx context.Context, topicSlug string, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)
	vars := map[string]interface{}{
		"topicSlug": topicSlug,
		"first":     o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}
	if o.order != "" {
		vars["order"] = string(o.order)
	}

	data, err := c.query(ctx, "TopicPostsQuery", topicPostsQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}
	return parsePostConnection(data)
}

// FollowTopic follows a topic.
func (c *Client) FollowTopic(ctx context.Context, topicID string) error {
	_, err := c.query(ctx, "TopicFollowMutation", topicFollowMutation, map[string]interface{}{
		"topicId": topicID,
	})
	return err
}

// UnfollowTopic unfollows a topic.
func (c *Client) UnfollowTopic(ctx context.Context, topicID string) error {
	_, err := c.query(ctx, "TopicUnfollowMutation", topicUnfollowMutation, map[string]interface{}{
		"topicId": topicID,
	})
	return err
}
