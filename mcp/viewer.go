package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// GetViewerInput is the typed input for producthunt_get_viewer.
type GetViewerInput struct{}

func getViewer(ctx context.Context, c *producthunt.Client, _ GetViewerInput) (any, error) {
	return c.GetViewer(ctx)
}

var viewerTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, GetViewerInput](
		"producthunt_get_viewer",
		"Fetch the currently authenticated Product Hunt user's profile",
		"GetViewer",
		getViewer,
	),
}
