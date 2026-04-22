package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetUserInput is the typed input for producthunt_get_user.
type GetUserInput struct {
	UsernameOrID string `json:"username_or_id" jsonschema:"description=Product Hunt username (e.g. 'rrhoover') or numeric user ID,required"`
}

func getUser(ctx context.Context, c *producthunt.Client, in GetUserInput) (any, error) {
	return c.GetUser(ctx, in.UsernameOrID)
}

// GetUserPostsInput is the typed input for producthunt_get_user_posts.
type GetUserPostsInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getUserPosts(ctx context.Context, c *producthunt.Client, in GetUserPostsInput) (any, error) {
	page, err := c.GetUserPosts(ctx, in.UserID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetUserVotedPostsInput is the typed input for producthunt_get_user_voted_posts.
type GetUserVotedPostsInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getUserVotedPosts(ctx context.Context, c *producthunt.Client, in GetUserVotedPostsInput) (any, error) {
	page, err := c.GetUserVotedPosts(ctx, in.UserID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetUserFollowersInput is the typed input for producthunt_get_user_followers.
type GetUserFollowersInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getUserFollowers(ctx context.Context, c *producthunt.Client, in GetUserFollowersInput) (any, error) {
	page, err := c.GetUserFollowers(ctx, in.UserID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// GetUserFollowingInput is the typed input for producthunt_get_user_following.
type GetUserFollowingInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getUserFollowing(ctx context.Context, c *producthunt.Client, in GetUserFollowingInput) (any, error) {
	page, err := c.GetUserFollowing(ctx, in.UserID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// FollowUserInput is the typed input for producthunt_follow_user.
type FollowUserInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID to follow,required"`
}

func followUser(ctx context.Context, c *producthunt.Client, in FollowUserInput) (any, error) {
	if err := c.FollowUser(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

// UnfollowUserInput is the typed input for producthunt_unfollow_user.
type UnfollowUserInput struct {
	UserID string `json:"user_id" jsonschema:"description=Product Hunt user ID to unfollow,required"`
}

func unfollowUser(ctx context.Context, c *producthunt.Client, in UnfollowUserInput) (any, error) {
	if err := c.UnfollowUser(ctx, in.UserID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "user_id": in.UserID}, nil
}

func pageOpts(first int, after string) []producthunt.PageOption {
	var opts []producthunt.PageOption
	if first > 0 {
		opts = append(opts, producthunt.WithFirst(first))
	}
	if after != "" {
		opts = append(opts, producthunt.WithAfter(after))
	}
	return opts
}

var usersTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetUserInput](
		"producthunt_get_user",
		"Fetch a Product Hunt user profile by username or numeric ID",
		"GetUser",
		getUser,
	),
	mcptool.Define[*producthunt.Client, GetUserPostsInput](
		"producthunt_get_user_posts",
		"List products a Product Hunt user has made (launched as a maker)",
		"GetUserPosts",
		getUserPosts,
	),
	mcptool.Define[*producthunt.Client, GetUserVotedPostsInput](
		"producthunt_get_user_voted_posts",
		"List products a Product Hunt user has upvoted",
		"GetUserVotedPosts",
		getUserVotedPosts,
	),
	mcptool.Define[*producthunt.Client, GetUserFollowersInput](
		"producthunt_get_user_followers",
		"List users following the given Product Hunt user",
		"GetUserFollowers",
		getUserFollowers,
	),
	mcptool.Define[*producthunt.Client, GetUserFollowingInput](
		"producthunt_get_user_following",
		"List users that the given Product Hunt user follows",
		"GetUserFollowing",
		getUserFollowing,
	),
	mcptool.Define[*producthunt.Client, FollowUserInput](
		"producthunt_follow_user",
		"Follow a Product Hunt user as the authenticated viewer",
		"FollowUser",
		followUser,
	),
	mcptool.Define[*producthunt.Client, UnfollowUserInput](
		"producthunt_unfollow_user",
		"Unfollow a Product Hunt user the authenticated viewer follows",
		"UnfollowUser",
		unfollowUser,
	),
}
