package producthunt

import (
	"context"
	"fmt"
)

const searchPostsQuery = `
query SearchQuery($query: String!, $first: Int, $after: String) {
  posts(url: $query, first: $first, after: $after) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { cursor node {` + postFields + `} }
  }
}
`

// Search searches Product Hunt by keyword across posts, users, and collections.
//
// Due to the public GraphQL API's search limitations (posts can only be
// filtered by url/topic/twitter, not free-text), this method implements search
// by querying the posts endpoint with the query as a url filter. For richer
// search, use GetTopicPosts with a topic slug, or GetPostsByDate with filters.
//
// When the internal frontend API is verified with proper cookies, this will be
// upgraded to use the full-text search endpoint.
func (c *Client) Search(ctx context.Context, query string, opts ...SearchOption) (*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: search query must be non-empty", ErrInvalidParams)
	}

	o := applySearchOptions(opts)
	result := &SearchResult{}

	if o.searchType == "" || o.searchType == SearchPosts {
		vars := map[string]interface{}{
			"query": query,
			"first": o.first,
		}
		if o.after != "" {
			vars["after"] = o.after
		}

		data, err := c.query(ctx, "SearchQuery", searchPostsQuery, vars)
		if err != nil {
			return nil, err
		}

		page, err := parsePostConnection(data)
		if err != nil {
			return nil, err
		}
		result.Posts = page
	}

	return result, nil
}
