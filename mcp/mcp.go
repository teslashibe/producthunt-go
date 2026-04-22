// Package mcp exposes the producthunt-go [producthunt.Client] surface as a
// set of MCP (Model Context Protocol) tools that any host application can
// mount on its own MCP server.
//
// All tools wrap exported methods on *producthunt.Client. Each tool is
// defined via [mcptool.Define] so the JSON input schema is reflected from the
// typed input struct — no hand-maintained schemas, no drift.
//
// Usage from a host application:
//
//	import (
//	    "github.com/teslashibe/mcptool"
//	    producthunt "github.com/teslashibe/producthunt-go"
//	    phmcp "github.com/teslashibe/producthunt-go/mcp"
//	)
//
//	client, _ := producthunt.New(producthunt.Credentials{...})
//	for _, tool := range (phmcp.Provider{}).Tools() {
//	    // register tool with your MCP server, passing client as the client
//	    // arg when invoking
//	}
//
// The [Excluded] map documents methods on *Client that are intentionally not
// exposed via MCP, with a one-line reason. The coverage test in mcp_test.go
// fails if a new exported method is added without either being wrapped by a
// tool or appearing in [Excluded].
package mcp

import "github.com/teslashibe/mcptool"

// Provider implements [mcptool.Provider] for producthunt-go. The zero value
// is ready to use.
type Provider struct{}

// Platform returns "producthunt".
func (Provider) Platform() string { return "producthunt" }

// Tools returns every producthunt-go MCP tool, in registration order.
func (Provider) Tools() []mcptool.Tool {
	out := make([]mcptool.Tool, 0,
		len(postsTools)+len(usersTools)+len(viewerTools)+
			len(topicsTools)+len(collectionsTools)+len(commentsTools)+
			len(reviewsTools)+len(votesTools)+len(searchTools)+
			len(trendsTools),
	)
	out = append(out, postsTools...)
	out = append(out, usersTools...)
	out = append(out, viewerTools...)
	out = append(out, topicsTools...)
	out = append(out, collectionsTools...)
	out = append(out, commentsTools...)
	out = append(out, reviewsTools...)
	out = append(out, votesTools...)
	out = append(out, searchTools...)
	out = append(out, trendsTools...)
	return out
}
