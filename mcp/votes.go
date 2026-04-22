package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetPostVotersInput is the typed input for producthunt_get_post_voters.
type GetPostVotersInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getPostVoters(ctx context.Context, c *producthunt.Client, in GetPostVotersInput) (any, error) {
	page, err := c.GetPostVoters(ctx, in.PostID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// UpvoteInput is the typed input for producthunt_upvote.
type UpvoteInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID to upvote,required"`
}

func upvote(ctx context.Context, c *producthunt.Client, in UpvoteInput) (any, error) {
	if err := c.Upvote(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

// RemoveUpvoteInput is the typed input for producthunt_remove_upvote.
type RemoveUpvoteInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID to remove the upvote from,required"`
}

func removeUpvote(ctx context.Context, c *producthunt.Client, in RemoveUpvoteInput) (any, error) {
	if err := c.RemoveUpvote(ctx, in.PostID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "post_id": in.PostID}, nil
}

var votesTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetPostVotersInput](
		"producthunt_get_post_voters",
		"List users who upvoted a Product Hunt post",
		"GetPostVoters",
		getPostVoters,
	),
	mcptool.Define[*producthunt.Client, UpvoteInput](
		"producthunt_upvote",
		"Upvote a Product Hunt post as the authenticated viewer",
		"Upvote",
		upvote,
	),
	mcptool.Define[*producthunt.Client, RemoveUpvoteInput](
		"producthunt_remove_upvote",
		"Remove the authenticated viewer's upvote from a Product Hunt post",
		"RemoveUpvote",
		removeUpvote,
	),
}
