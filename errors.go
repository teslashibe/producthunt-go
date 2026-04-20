package producthunt

import "errors"

var (
	ErrInvalidAuth   = errors.New("producthunt: missing or empty session cookie")
	ErrUnauthorized  = errors.New("producthunt: authentication failed (session expired or invalid)")
	ErrCFChallenge   = errors.New("producthunt: Cloudflare challenge required — refresh cf_clearance cookie")
	ErrForbidden     = errors.New("producthunt: access denied")
	ErrNotFound      = errors.New("producthunt: resource not found")
	ErrRateLimited   = errors.New("producthunt: rate limited")
	ErrAlreadyVoted  = errors.New("producthunt: already upvoted this post")
	ErrNotVoted      = errors.New("producthunt: not voted on this post")
	ErrInvalidParams = errors.New("producthunt: invalid or missing required parameters")
	ErrPartialResult = errors.New("producthunt: context cancelled; partial result returned")
	ErrGraphQL       = errors.New("producthunt: GraphQL error")
	ErrRequestFailed = errors.New("producthunt: HTTP request failed")
)
