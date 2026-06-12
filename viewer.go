package producthunt

import "context"

const viewerQuery = `
query ViewerQuery {
  viewer {
    user {
      id
      username
      name
      headline
      profileImage
      coverImage
      twitterUsername
      websiteUrl
      url
      followersCount
      followingCount
      isMaker
      createdAt
    }
  }
}
`

// GetViewer returns the currently authenticated user.
func (c *Client) GetViewer(ctx context.Context) (*User, error) {
	data, err := c.query(ctx, "ViewerQuery", viewerQuery, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Viewer struct {
			User gqlUser `json:"user"`
		} `json:"viewer"`
	}
	if err := unmarshalData(data, &resp); err != nil {
		return nil, err
	}

	u := resp.Viewer.User.toUser()
	u.IsViewer = true
	return &u, nil
}

// GetMe is a liveness alias for GetViewer. Hosts probe a credential's health
// by calling GetMe(ctx) (the smore liveness prober adapts any
// GetMe(ctx) (..., error) method into a HealthCheck), so this lets a stored
// Product Hunt developer token be verified as live, not merely present.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	return c.GetViewer(ctx)
}
