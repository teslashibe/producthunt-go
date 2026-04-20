package producthunt

import (
	"context"
	"fmt"
)

const commentFields = `
  id
  body
  votesCount
  isVoted
  user { id username name headline profileImage url }
  parentId
  createdAt
  replies(first: 10, order: NEWEST) {
    edges { node {
      id body votesCount isVoted createdAt parentId
      user { id username name headline profileImage url }
    } }
  }
`

const postCommentsQuery = `
query PostCommentsQuery($postId: ID!, $order: CommentsOrder, $first: Int, $after: String) {
  post(id: $postId) {
    comments(order: $order, first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + commentFields + `} }
    }
  }
}
`

const commentCreateMutation = `
mutation CommentCreateMutation($postId: ID!, $body: String!, $parentId: ID) {
  commentCreate(input: { postId: $postId, body: $body, parentId: $parentId }) {
    comment {` + commentFields + `}
  }
}
`

const commentUpdateMutation = `
mutation CommentUpdateMutation($commentId: ID!, $body: String!) {
  commentUpdate(input: { id: $commentId, body: $body }) {
    comment {` + commentFields + `}
  }
}
`

const commentDeleteMutation = `
mutation CommentDeleteMutation($commentId: ID!) {
  commentDelete(input: { id: $commentId }) {
    comment { id }
  }
}
`

// GetPostComments returns comments on a post.
func (c *Client) GetPostComments(ctx context.Context, postID string, opts ...CommentListOption) (Page[Comment], error) {
	o := applyCommentListOptions(opts)
	vars := map[string]interface{}{
		"postId": postID,
		"order":  string(o.order),
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "PostCommentsQuery", postCommentsQuery, vars)
	if err != nil {
		return Page[Comment]{}, err
	}

	var resp struct {
		Post *struct {
			Comments relayConnection[gqlComment] `json:"comments"`
		} `json:"post"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Comment]{}, err
	}
	if resp.Post == nil {
		return Page[Comment]{}, ErrNotFound
	}

	conn := resp.Post.Comments
	page := Page[Comment]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]Comment, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toComment())
	}
	return page, nil
}

// CreateComment posts a top-level comment on a product.
func (c *Client) CreateComment(ctx context.Context, postID, body string) (*Comment, error) {
	if body == "" {
		return nil, fmt.Errorf("%w: comment body must be non-empty", ErrInvalidParams)
	}
	vars := map[string]interface{}{
		"postId": postID,
		"body":   body,
	}
	return c.doCreateComment(ctx, vars)
}

// ReplyToComment creates a threaded reply to an existing comment.
func (c *Client) ReplyToComment(ctx context.Context, postID, parentCommentID, body string) (*Comment, error) {
	if body == "" {
		return nil, fmt.Errorf("%w: reply body must be non-empty", ErrInvalidParams)
	}
	vars := map[string]interface{}{
		"postId":   postID,
		"body":     body,
		"parentId": parentCommentID,
	}
	return c.doCreateComment(ctx, vars)
}

func (c *Client) doCreateComment(ctx context.Context, vars map[string]interface{}) (*Comment, error) {
	data, err := c.query(ctx, "CommentCreateMutation", commentCreateMutation, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		CommentCreate *struct {
			Comment gqlComment `json:"comment"`
		} `json:"commentCreate"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.CommentCreate == nil {
		return nil, fmt.Errorf("%w: no comment returned", ErrRequestFailed)
	}

	comment := resp.CommentCreate.Comment.toComment()
	return &comment, nil
}

// UpdateComment edits the caller's own comment.
func (c *Client) UpdateComment(ctx context.Context, commentID, body string) (*Comment, error) {
	if body == "" {
		return nil, fmt.Errorf("%w: comment body must be non-empty", ErrInvalidParams)
	}
	vars := map[string]interface{}{
		"commentId": commentID,
		"body":      body,
	}

	data, err := c.query(ctx, "CommentUpdateMutation", commentUpdateMutation, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		CommentUpdate *struct {
			Comment gqlComment `json:"comment"`
		} `json:"commentUpdate"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.CommentUpdate == nil {
		return nil, fmt.Errorf("%w: no comment returned", ErrRequestFailed)
	}

	comment := resp.CommentUpdate.Comment.toComment()
	return &comment, nil
}

// DeleteComment removes the caller's own comment.
func (c *Client) DeleteComment(ctx context.Context, commentID string) error {
	vars := map[string]interface{}{
		"commentId": commentID,
	}
	_, err := c.query(ctx, "CommentDeleteMutation", commentDeleteMutation, vars)
	return err
}
