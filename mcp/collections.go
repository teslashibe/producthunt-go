package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetCollectionsInput is the typed input for producthunt_get_collections.
//
// Note: the underlying client does not expose page-size or cursor setters
// for collection listings, so results are capped at the default of 20 per
// call.
type GetCollectionsInput struct {
	Featured *bool  `json:"featured,omitempty" jsonschema:"description=filter to featured collections only when set"`
	UserID   string `json:"user_id,omitempty" jsonschema:"description=filter collections by curator user ID"`
	PostID   string `json:"post_id,omitempty" jsonschema:"description=filter collections that contain the given post ID"`
	Order    string `json:"order,omitempty" jsonschema:"description=sort order; allowed: FEATURED_AT,FOLLOWERS_COUNT,NEWEST"`
}

func getCollections(ctx context.Context, c *producthunt.Client, in GetCollectionsInput) (any, error) {
	var opts []producthunt.CollectionListOption
	if in.Order != "" {
		opts = append(opts, producthunt.WithCollectionsOrder(producthunt.CollectionsOrder(in.Order)))
	}
	if in.Featured != nil {
		opts = append(opts, producthunt.WithCollectionsFeatured(*in.Featured))
	}
	if in.UserID != "" {
		opts = append(opts, producthunt.WithCollectionsUserID(in.UserID))
	}
	if in.PostID != "" {
		opts = append(opts, producthunt.WithCollectionsPostID(in.PostID))
	}
	page, err := c.GetCollections(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, 0), nil
}

// GetCollectionInput is the typed input for producthunt_get_collection.
type GetCollectionInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID,required"`
}

func getCollection(ctx context.Context, c *producthunt.Client, in GetCollectionInput) (any, error) {
	return c.GetCollection(ctx, in.CollectionID)
}

// GetCollectionPostsInput is the typed input for producthunt_get_collection_posts.
type GetCollectionPostsInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID,required"`
	First        int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After        string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getCollectionPosts(ctx context.Context, c *producthunt.Client, in GetCollectionPostsInput) (any, error) {
	opts := buildPostListOpts(in.First, in.After, "", nil, "")
	page, err := c.GetCollectionPosts(ctx, in.CollectionID, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// CreateCollectionInput is the typed input for producthunt_create_collection.
type CreateCollectionInput struct {
	Name        string `json:"name" jsonschema:"description=collection name,required"`
	Tagline     string `json:"tagline" jsonschema:"description=short collection tagline,required"`
	Description string `json:"description,omitempty" jsonschema:"description=longer collection description"`
}

func createCollection(ctx context.Context, c *producthunt.Client, in CreateCollectionInput) (any, error) {
	col, err := c.CreateCollection(ctx, producthunt.CreateCollectionParams{
		Name:        in.Name,
		Tagline:     in.Tagline,
		Description: in.Description,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "collection": col}, nil
}

// AddPostToCollectionInput is the typed input for producthunt_add_post_to_collection.
type AddPostToCollectionInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID,required"`
	PostID       string `json:"post_id" jsonschema:"description=Product Hunt post ID to add,required"`
}

func addPostToCollection(ctx context.Context, c *producthunt.Client, in AddPostToCollectionInput) (any, error) {
	if err := c.AddPostToCollection(ctx, in.CollectionID, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "collection_id": in.CollectionID, "post_id": in.PostID}, nil
}

// RemovePostFromCollectionInput is the typed input for producthunt_remove_post_from_collection.
type RemovePostFromCollectionInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID,required"`
	PostID       string `json:"post_id" jsonschema:"description=Product Hunt post ID to remove,required"`
}

func removePostFromCollection(ctx context.Context, c *producthunt.Client, in RemovePostFromCollectionInput) (any, error) {
	if err := c.RemovePostFromCollection(ctx, in.CollectionID, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "collection_id": in.CollectionID, "post_id": in.PostID}, nil
}

// FollowCollectionInput is the typed input for producthunt_follow_collection.
type FollowCollectionInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID to follow,required"`
}

func followCollection(ctx context.Context, c *producthunt.Client, in FollowCollectionInput) (any, error) {
	if err := c.FollowCollection(ctx, in.CollectionID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "collection_id": in.CollectionID}, nil
}

// UnfollowCollectionInput is the typed input for producthunt_unfollow_collection.
type UnfollowCollectionInput struct {
	CollectionID string `json:"collection_id" jsonschema:"description=Product Hunt collection ID to unfollow,required"`
}

func unfollowCollection(ctx context.Context, c *producthunt.Client, in UnfollowCollectionInput) (any, error) {
	if err := c.UnfollowCollection(ctx, in.CollectionID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "collection_id": in.CollectionID}, nil
}

var collectionsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetCollectionsInput](
		"producthunt_get_collections",
		"List Product Hunt collections with optional featured/user/post/order filters",
		"GetCollections",
		getCollections,
	),
	mcptool.Define[*producthunt.Client, GetCollectionInput](
		"producthunt_get_collection",
		"Fetch a single Product Hunt collection by ID",
		"GetCollection",
		getCollection,
	),
	mcptool.Define[*producthunt.Client, GetCollectionPostsInput](
		"producthunt_get_collection_posts",
		"List posts within a Product Hunt collection",
		"GetCollectionPosts",
		getCollectionPosts,
	),
	mcptool.Define[*producthunt.Client, CreateCollectionInput](
		"producthunt_create_collection",
		"Create a new Product Hunt collection (requires authenticated viewer)",
		"CreateCollection",
		createCollection,
	),
	mcptool.Define[*producthunt.Client, AddPostToCollectionInput](
		"producthunt_add_post_to_collection",
		"Add a post to a Product Hunt collection owned by the viewer",
		"AddPostToCollection",
		addPostToCollection,
	),
	mcptool.Define[*producthunt.Client, RemovePostFromCollectionInput](
		"producthunt_remove_post_from_collection",
		"Remove a post from a Product Hunt collection owned by the viewer",
		"RemovePostFromCollection",
		removePostFromCollection,
	),
	mcptool.Define[*producthunt.Client, FollowCollectionInput](
		"producthunt_follow_collection",
		"Follow a Product Hunt collection as the authenticated viewer",
		"FollowCollection",
		followCollection,
	),
	mcptool.Define[*producthunt.Client, UnfollowCollectionInput](
		"producthunt_unfollow_collection",
		"Unfollow a Product Hunt collection the authenticated viewer follows",
		"UnfollowCollection",
		unfollowCollection,
	),
}
