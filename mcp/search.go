package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// SearchInput is the typed input for producthunt_search.
type SearchInput struct {
	Query string `json:"query" jsonschema:"description=keywords or URL to search Product Hunt for,required"`
	Type  string `json:"type,omitempty" jsonschema:"description=restrict results; allowed: POSTS,USERS,COLLECTIONS,default=POSTS"`
	First int    `json:"first,omitempty" jsonschema:"description=results per page,minimum=1,maximum=50,default=20"`
	After string `json:"after,omitempty" jsonschema:"description=pagination cursor"`
}

func search(ctx context.Context, c *producthunt.Client, in SearchInput) (any, error) {
	var opts []producthunt.SearchOption
	if in.Type != "" {
		opts = append(opts, producthunt.WithSearchType(producthunt.SearchType(in.Type)))
	}
	if in.First > 0 {
		opts = append(opts, producthunt.WithSearchFirst(in.First))
	}
	if in.After != "" {
		opts = append(opts, producthunt.WithSearchAfter(in.After))
	}
	return c.Search(ctx, in.Query, opts...)
}

var searchTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, SearchInput](
		"producthunt_search",
		"Search Product Hunt for posts (and, with type, users or collections) by keyword or URL",
		"Search",
		search,
	),
}
