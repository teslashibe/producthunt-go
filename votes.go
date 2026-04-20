package producthunt

import "context"

const postVotersQuery = `
query PostVotersQuery($postId: ID!, $first: Int, $after: String) {
  post(id: $postId) {
    votes(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {
        id
        createdAt
        user { id username name headline profileImage url }
      } }
    }
  }
}
`

const voteCreateMutation = `
mutation VoteCreateMutation($postId: ID!) {
  voteCreate(input: { postId: $postId }) {
    node { id votesCount isVoted }
  }
}
`

const voteDeleteMutation = `
mutation VoteDeleteMutation($postId: ID!) {
  voteDelete(input: { postId: $postId }) {
    node { id votesCount isVoted }
  }
}
`

// GetPostVoters returns a paginated list of users who upvoted a post.
func (c *Client) GetPostVoters(ctx context.Context, postID string, opts ...PageOption) (Page[Vote], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"postId": postID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "PostVotersQuery", postVotersQuery, vars)
	if err != nil {
		return Page[Vote]{}, err
	}

	var resp struct {
		Post *struct {
			Votes relayConnection[gqlVote] `json:"votes"`
		} `json:"post"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Vote]{}, err
	}
	if resp.Post == nil {
		return Page[Vote]{}, ErrNotFound
	}

	conn := resp.Post.Votes
	page := Page[Vote]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]Vote, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toVote())
	}
	return page, nil
}

// Upvote upvotes a post.
func (c *Client) Upvote(ctx context.Context, postID string) error {
	vars := map[string]interface{}{"postId": postID}
	_, err := c.query(ctx, "VoteCreateMutation", voteCreateMutation, vars)
	return err
}

// RemoveUpvote removes an upvote from a post.
func (c *Client) RemoveUpvote(ctx context.Context, postID string) error {
	vars := map[string]interface{}{"postId": postID}
	_, err := c.query(ctx, "VoteDeleteMutation", voteDeleteMutation, vars)
	return err
}
