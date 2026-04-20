package producthunt

import (
	"context"
	"fmt"
)

const userFields = `
  id username name headline profileImage coverImage twitterUsername websiteUrl
  url followersCount followingCount isMaker isFollowing isViewer createdAt
`

const singleUserByUsernameQuery = `
query UserQuery($username: String!) {
  user(username: $username) {` + userFields + `}
}
`

const singleUserByIDQuery = `
query UserQuery($id: ID!) {
  user(id: $id) {` + userFields + `}
}
`

const userMadePostsQuery = `
query UserPostsQuery($userId: ID!, $first: Int, $after: String) {
  user(id: $userId) {
    madePosts(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + postFields + `} }
    }
  }
}
`

const userVotedPostsQuery = `
query UserVotedPostsQuery($userId: ID!, $first: Int, $after: String) {
  user(id: $userId) {
    votedPosts(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + postFields + `} }
    }
  }
}
`

const userFollowersQuery = `
query UserFollowersQuery($userId: ID!, $first: Int, $after: String) {
  user(id: $userId) {
    followers(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + userFields + `} }
    }
  }
}
`

const userFollowingQuery = `
query UserFollowingQuery($userId: ID!, $first: Int, $after: String) {
  user(id: $userId) {
    following(first: $first, after: $after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { cursor node {` + userFields + `} }
    }
  }
}
`

const userFollowMutation = `
mutation UserFollowMutation($userId: ID!) {
  userFollow(input: { userId: $userId }) {
    node { id isFollowing }
  }
}
`

const userUnfollowMutation = `
mutation UserUnfollowMutation($userId: ID!) {
  userFollowUndo(input: { userId: $userId }) {
    node { id isFollowing }
  }
}
`

// GetUser returns a user profile by username or ID.
func (c *Client) GetUser(ctx context.Context, usernameOrID string) (*User, error) {
	if usernameOrID == "" {
		return nil, fmt.Errorf("%w: usernameOrID must be non-empty", ErrInvalidParams)
	}

	var q string
	var vars map[string]interface{}

	if isNumeric(usernameOrID) {
		q = singleUserByIDQuery
		vars = map[string]interface{}{"id": usernameOrID}
	} else {
		q = singleUserByUsernameQuery
		vars = map[string]interface{}{"username": usernameOrID}
	}

	data, err := c.query(ctx, "UserQuery", q, vars)
	if err != nil {
		return nil, err
	}

	var resp struct {
		User *gqlUser `json:"user"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}
	if resp.User == nil {
		return nil, ErrNotFound
	}

	u := resp.User.toUser()
	return &u, nil
}

// GetUserPosts returns products the user has made.
func (c *Client) GetUserPosts(ctx context.Context, userID string, opts ...PageOption) (Page[Post], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"userId": userID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "UserPostsQuery", userMadePostsQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}

	var resp struct {
		User *struct {
			MadePosts relayConnection[gqlPost] `json:"madePosts"`
		} `json:"user"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Post]{}, err
	}
	if resp.User == nil {
		return Page[Post]{}, ErrNotFound
	}

	return convertPostConnection(resp.User.MadePosts), nil
}

// GetUserVotedPosts returns products the user has upvoted.
func (c *Client) GetUserVotedPosts(ctx context.Context, userID string, opts ...PageOption) (Page[Post], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"userId": userID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "UserVotedPostsQuery", userVotedPostsQuery, vars)
	if err != nil {
		return Page[Post]{}, err
	}

	var resp struct {
		User *struct {
			VotedPosts relayConnection[gqlPost] `json:"votedPosts"`
		} `json:"user"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[Post]{}, err
	}
	if resp.User == nil {
		return Page[Post]{}, ErrNotFound
	}

	return convertPostConnection(resp.User.VotedPosts), nil
}

// GetUserFollowers returns users who follow the given user.
func (c *Client) GetUserFollowers(ctx context.Context, userID string, opts ...PageOption) (Page[User], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"userId": userID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "UserFollowersQuery", userFollowersQuery, vars)
	if err != nil {
		return Page[User]{}, err
	}

	var resp struct {
		User *struct {
			Followers relayConnection[gqlUser] `json:"followers"`
		} `json:"user"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[User]{}, err
	}
	if resp.User == nil {
		return Page[User]{}, ErrNotFound
	}

	return convertUserConnection(resp.User.Followers), nil
}

// GetUserFollowing returns users that the given user follows.
func (c *Client) GetUserFollowing(ctx context.Context, userID string, opts ...PageOption) (Page[User], error) {
	o := applyPageOptions(opts)
	vars := map[string]interface{}{
		"userId": userID,
		"first":  o.first,
	}
	if o.after != "" {
		vars["after"] = o.after
	}

	data, err := c.query(ctx, "UserFollowingQuery", userFollowingQuery, vars)
	if err != nil {
		return Page[User]{}, err
	}

	var resp struct {
		User *struct {
			Following relayConnection[gqlUser] `json:"following"`
		} `json:"user"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return Page[User]{}, err
	}
	if resp.User == nil {
		return Page[User]{}, ErrNotFound
	}

	return convertUserConnection(resp.User.Following), nil
}

// FollowUser follows a user.
func (c *Client) FollowUser(ctx context.Context, userID string) error {
	_, err := c.query(ctx, "UserFollowMutation", userFollowMutation, map[string]interface{}{
		"userId": userID,
	})
	return err
}

// UnfollowUser unfollows a user.
func (c *Client) UnfollowUser(ctx context.Context, userID string) error {
	_, err := c.query(ctx, "UserUnfollowMutation", userUnfollowMutation, map[string]interface{}{
		"userId": userID,
	})
	return err
}

// --- helpers ---

func convertPostConnection(conn relayConnection[gqlPost]) Page[Post] {
	page := Page[Post]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]Post, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toPost())
	}
	return page
}

func convertUserConnection(conn relayConnection[gqlUser]) Page[User] {
	page := Page[User]{
		TotalCount: conn.TotalCount,
		HasNext:    conn.PageInfo.HasNextPage,
		EndCursor:  conn.PageInfo.EndCursor,
	}
	page.Items = make([]User, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		page.Items = append(page.Items, e.Node.toUser())
	}
	return page
}
