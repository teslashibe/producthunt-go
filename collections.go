package producthunt

import (
	"context"
	"fmt"
)

const collectionFields = `
  id name tagline description coverImage followersCount isFollowing url createdAt featuredAt
  user { id username name headline profileImage url }
`

const collectionsListQuery = `
query CollectionsQuery($featured: Boolean, $userId: ID, $postId: ID, $order: CollectionsOrder, $first: Int, $after: String) {
  collections(featured: $featured, userId: $userId, postId: $postId, order: $order, first: $first, after: $after) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + collectionFields + `} }
  }
}
`

const singleCollectionQuery = `
query CollectionQuery($id: ID!) {
  collection(id: $id) {` + collectionFields + `}
}
`

const collectionPostsQuery = `
query CollectionPostsQuery($collectionId: ID!, $first: Int, $after: String) {
  collection(id: $collectionId) {
    posts(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + postFields + `} }
    }
  }
}
`

const collectionCreateMutation = `
mutation CollectionCreateMutation($name: String!, $tagline: String!, $description: String) {
  collectionCreate(input: { name: $name, tagline: $tagline, description: $description }) {
    collection {` + collectionFields + `}
  }
}
`

const collectionAddPostMutation = `
mutation CollectionAddPostMutation($collectionId: ID!, $postId: ID!) {
  collectionAddPost(input: { collectionId: $collectionId, postId: $postId }) {
    collection { id }
  }
}
`

const collectionRemovePostMutation = `
mutation CollectionRemovePostMutation($collectionId: ID!, $postId: ID!) {
  collectionRemovePost(input: { collectionId: $collectionId, postId: $postId }) {
    collection { id }
  }
}
`

const collectionFollowMutation = `
mutation CollectionFollowMutation($collectionId: ID!) {
  collectionFollow(input: { collectionId: $collectionId }) {
    collection { id isFollowing }
  }
}
`

const collectionUnfollowMutation = `
mutation CollectionUnfollowMutation($collectionId: ID!) {
  collectionUnfollow(input: { collectionId: $collectionId }) {
    collection { id isFollowing }
  }
}
`

// GetCollections returns a paginated list of collections.
func (c *Client) GetCollections(ctx context.Context, opts ...CollectionListOption) (Page[Collection], error) {
	o := applyCollectionListOptions(opts)
	vars := map[string]interface{}{
		"first": o.first,
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
	if o.userID != "" {
		vars["userId"] = o.userID
	}
	if o.postID != "" {
		vars["postId"] = o.postID
	}

	data, err := c.query(ctx, "CollectionsQuery", collectionsListQuery, vars)
	if err != nil {
		return Page[Collection]{}, err
	}

	var resp struct {
		Collections relayConnection[gqlCollection] `json:"collections"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Collection]{}, err
	}

	page := Page[Collection]{
		TotalCount: resp.Collections.TotalCount,
		HasNext:    resp.Collections.PageInfo.HasNextPage,
		EndCursor:  resp.Collections.PageInfo.EndCursor,
	}
	page.Items = make([]Collection, 0, len(resp.Collections.Edges))
	for _, e := range resp.Collections.Edges {
		page.Items = append(page.Items, e.Node.toCollection())
	}
	return page, nil
}

// GetCollection returns a single collection by ID.
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*Collection, error) {
	data, err := c.query(ctx, "CollectionQuery", singleCollectionQuery, map[string]interface{}{
		"id": collectionID,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Collection *gqlCollection `json:"collection"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.Collection == nil {
		return nil, ErrNotFound
	}

	col := resp.Collection.toCollection()
	return &col, nil
}

// GetCollectionPosts returns posts within a collection.
func (c *Client) GetCollectionPosts(ctx context.Context, collectionID string, opts ...PostListOption) (Page[Post], error) {
	o := applyPostListOptions(opts)
	vars := map[string]interface{}{
		"collectionId": collectionID,
		"first":        o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "CollectionPostsQuery", collectionPostsQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}

	var resp struct {
		Collection *struct {
			Posts relayConnection[gqlPost] `json:"posts"`
		} `json:"collection"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Post]{}, err
	}
	if resp.Collection == nil {
		return Page[Post]{}, ErrNotFound
	}

	conn := resp.Collection.Posts
	page := Page[Post]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]Post, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toPost())
	}
	return page, nil
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, params CreateCollectionParams) (*Collection, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: collection name must be non-empty", ErrInvalidParams)
	}
	if params.Tagline == "" {
		return nil, fmt.Errorf("%w: collection tagline must be non-empty", ErrInvalidParams)
	}

	vars := map[string]interface{}{
		"name":    params.Name,
		"tagline": params.Tagline,
	}
	if params.Description != "" {
		vars["description"] = params.Description
	}

	data, err := c.query(ctx, "CollectionCreateMutation", collectionCreateMutation, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		CollectionCreate *struct {
			Collection gqlCollection `json:"collection"`
		} `json:"collectionCreate"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.CollectionCreate == nil {
		return nil, fmt.Errorf("%w: no collection returned", ErrRequestFailed)
	}

	col := resp.CollectionCreate.Collection.toCollection()
	return &col, nil
}

// AddPostToCollection adds a post to a collection.
func (c *Client) AddPostToCollection(ctx context.Context, collectionID, postID string) error {
	_, err := c.query(ctx, "CollectionAddPostMutation", collectionAddPostMutation, map[string]interface{}{
		"collectionId": collectionID,
		"postId":       postID,
	})
	return err
}

// RemovePostFromCollection removes a post from a collection.
func (c *Client) RemovePostFromCollection(ctx context.Context, collectionID, postID string) error {
	_, err := c.query(ctx, "CollectionRemovePostMutation", collectionRemovePostMutation, map[string]interface{}{
		"collectionId": collectionID,
		"postId":       postID,
	})
	return err
}

// FollowCollection follows a collection.
func (c *Client) FollowCollection(ctx context.Context, collectionID string) error {
	_, err := c.query(ctx, "CollectionFollowMutation", collectionFollowMutation, map[string]interface{}{
		"collectionId": collectionID,
	})
	return err
}

// UnfollowCollection unfollows a collection.
func (c *Client) UnfollowCollection(ctx context.Context, collectionID string) error {
	_, err := c.query(ctx, "CollectionUnfollowMutation", collectionUnfollowMutation, map[string]interface{}{
		"collectionId": collectionID,
	})
	return err
}
