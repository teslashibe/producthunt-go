package producthunt

import (
	"context"
	"fmt"
)

const reviewFields = `
  id
  body
  rating
  sentiment
  user { id username name headline profileImage url }
  createdAt
`

const postReviewsQuery = `
query PostReviewsQuery($postId: ID!, $first: Int, $after: String) {
  post(id: $postId) {
    reviews(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + reviewFields + `} }
    }
  }
}
`

const reviewCreateMutation = `
mutation ReviewCreateMutation($postId: ID!, $body: String!, $rating: Float!) {
  reviewCreate(input: { postId: $postId, body: $body, rating: $rating }) {
    review {` + reviewFields + `}
  }
}
`

// GetPostReviews returns reviews for a post.
func (c *Client) GetPostReviews(ctx context.Context, postID string, opts ...PageOption) (Page[Review], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"postId": postID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "PostReviewsQuery", postReviewsQuery, vars)
	if err != nil {
		return Page[Review]{}, err
	}

	var resp struct {
		Post *struct {
			Reviews relayConnection[gqlReview] `json:"reviews"`
		} `json:"post"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Review]{}, err
	}
	if resp.Post == nil {
		return Page[Review]{}, ErrNotFound
	}

	conn := resp.Post.Reviews
	page := Page[Review]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]Review, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toReview())
	}
	return page, nil
}

// CreateReview writes a review for a product. Rating must be between 1.0 and 5.0.
func (c *Client) CreateReview(ctx context.Context, postID string, body string, rating float64) (*Review, error) {
	if rating < 1.0 || rating > 5.0 {
		return nil, fmt.Errorf("%w: rating must be between 1.0 and 5.0", ErrInvalidParams)
	}
	if body == "" {
		return nil, fmt.Errorf("%w: review body must be non-empty", ErrInvalidParams)
	}

	vars := map[string]interface{}{
		"postId": postID,
		"body":   body,
		"rating": rating,
	}

	data, err := c.query(ctx, "ReviewCreateMutation", reviewCreateMutation, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ReviewCreate *struct {
			Review gqlReview `json:"review"`
		} `json:"reviewCreate"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.ReviewCreate == nil {
		return nil, fmt.Errorf("%w: no review returned", ErrRequestFailed)
	}

	r := resp.ReviewCreate.Review.toReview()
	return &r, nil
}
