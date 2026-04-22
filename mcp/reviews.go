package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetPostReviewsInput is the typed input for producthunt_get_post_reviews.
type GetPostReviewsInput struct {
	PostID string `json:"post_id" jsonschema:"description=Product Hunt post ID,required"`
	First  int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After  string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func getPostReviews(ctx context.Context, c *producthunt.Client, in GetPostReviewsInput) (any, error) {
	page, err := c.GetPostReviews(ctx, in.PostID, pageOpts(in.First, in.After)...)
	if err != nil {
		return nil, err
	}
	return mcptool.PageOf(page.Items, page.EndCursor, in.First), nil
}

// CreateReviewInput is the typed input for producthunt_create_review.
type CreateReviewInput struct {
	PostID string  `json:"post_id" jsonschema:"description=Product Hunt post ID to review,required"`
	Body   string  `json:"body" jsonschema:"description=plain-text review body,required"`
	Rating float64 `json:"rating" jsonschema:"description=star rating between 1.0 and 5.0,minimum=1,maximum=5,required"`
}

func createReview(ctx context.Context, c *producthunt.Client, in CreateReviewInput) (any, error) {
	r, err := c.CreateReview(ctx, in.PostID, in.Body, in.Rating)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "review": r}, nil
}

var reviewsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetPostReviewsInput](
		"producthunt_get_post_reviews",
		"List reviews for a Product Hunt post (1.0-5.0 ratings with body and sentiment)",
		"GetPostReviews",
		getPostReviews,
	),
	mcptool.Define[*producthunt.Client, CreateReviewInput](
		"producthunt_create_review",
		"Write a Product Hunt review with body and 1.0-5.0 star rating",
		"CreateReview",
		createReview,
	),
}
