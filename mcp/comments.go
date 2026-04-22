package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetPostCommentsInput is the typed input for producthunt_get_post_comments.
//
// Note: the underlying client does not expose page-size or cursor setters
// for comment listings, so results are capped at the default of 20 per call.
type GetPostCommentsInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID,required"`
	Order  string `json:"order,omitempty" jsonschema:"description=sort order; allowed: NEWEST,OLDEST,default=NEWEST"`
}

func getPostComments(ctx context.Context, c *producthunt.Client, in GetPostCommentsInput) (any, error) {
	var opts []producthunt.CommentListOption
	if in.Order != "" {
		opts = append(opts, producthunt.WithCommentsOrder(producthunt.CommentsOrder(in.Order)))
	}
	page, err := c.GetPostComments(ctx, in.PostID, opts...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, 0), nil
}

// CreateCommentInput is the typed input for producthunt_create_comment.
type CreateCommentInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID to comment on,required"`
	Body   string `json:"body" jsonschema:"description=plain-text comment body,required"`
}

func createComment(ctx context.Context, c *producthunt.Client, in CreateCommentInput) (any, error) {
	cm, err := c.CreateComment(ctx, in.PostID, in.Body)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment": cm}, nil
}

// ReplyToCommentInput is the typed input for producthunt_reply_to_comment.
type ReplyToCommentInput struct {
	PostID          string `json:"post_id" jsonschema:"description=Product Hunt post ID hosting the parent comment,required"`
	ParentCommentID string `json:"parent_comment_id" jsonschema:"description=parent comment ID to reply to,required"`
	Body            string `json:"body" jsonschema:"description=plain-text reply body,required"`
}

func replyToComment(ctx context.Context, c *producthunt.Client, in ReplyToCommentInput) (any, error) {
	cm, err := c.ReplyToComment(ctx, in.PostID, in.ParentCommentID, in.Body)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment": cm}, nil
}

// UpdateCommentInput is the typed input for producthunt_update_comment.
type UpdateCommentInput struct {
	CommentID string `json:"comment_id" jsonschema:"description=Product Hunt comment ID owned by the viewer,required"`
	Body      string `json:"body" jsonschema:"description=updated plain-text body,required"`
}

func updateComment(ctx context.Context, c *producthunt.Client, in UpdateCommentInput) (any, error) {
	cm, err := c.UpdateComment(ctx, in.CommentID, in.Body)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment": cm}, nil
}

// DeleteCommentInput is the typed input for producthunt_delete_comment.
type DeleteCommentInput struct {
	CommentID string `json:"comment_id" jsonschema:"description=Product Hunt comment ID owned by the viewer,required"`
}

func deleteComment(ctx context.Context, c *producthunt.Client, in DeleteCommentInput) (any, error) {
	if err := c.DeleteComment(ctx, in.CommentID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "comment_id": in.CommentID}, nil
}

var commentsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetPostCommentsInput](
		"producthunt_get_post_comments",
		"List threaded comments on a Product Hunt post",
		"GetPostComments",
		getPostComments,
	),
	mcptool.Define[*producthunt.Client, CreateCommentInput](
		"producthunt_create_comment",
		"Post a top-level comment on a Product Hunt product",
		"CreateComment",
		createComment,
	),
	mcptool.Define[*producthunt.Client, ReplyToCommentInput](
		"producthunt_reply_to_comment",
		"Post a threaded reply to an existing Product Hunt comment",
		"ReplyToComment",
		replyToComment,
	),
	mcptool.Define[*producthunt.Client, UpdateCommentInput](
		"producthunt_update_comment",
		"Edit a comment owned by the authenticated viewer",
		"UpdateComment",
		updateComment,
	),
	mcptool.Define[*producthunt.Client, DeleteCommentInput](
		"producthunt_delete_comment",
		"Delete a comment owned by the authenticated viewer",
		"DeleteComment",
		deleteComment,
	),
}
