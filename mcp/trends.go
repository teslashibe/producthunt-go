package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	producthunt "github.com/teslashibe/producthunt-go"
)

// AnalyzeTrendsInput is the typed input for producthunt_analyze_trends.
type AnalyzeTrendsInput struct {
	From      string   `json:"from,omitempty" jsonschema:"description=start of analysis window in YYYY-MM-DD or RFC3339; defaults to 7 days ago"`
	To        string   `json:"to,omitempty" jsonschema:"description=end of analysis window in YYYY-MM-DD or RFC3339; defaults to now"`
	Topics    []string `json:"topics,omitempty" jsonschema:"description=topic slugs to scope the analysis to"`
	TopN      int      `json:"top_n,omitempty" jsonschema:"description=number of top items returned per dimension,minimum=1,maximum=100,default=20"`
	StopWords []string `json:"stop_words,omitempty" jsonschema:"description=extra stop words to exclude from keyword extraction"`
}

func analyzeTrends(ctx context.Context, c *producthunt.Client, in AnalyzeTrendsInput) (any, error) {
	var opts []producthunt.TrendOption
	if in.From != "" && in.To != "" {
		from, err := parseDate(in.From)
		if err != nil {
			return nil, err
		}
		to, err := parseDate(in.To)
		if err != nil {
			return nil, err
		}
		opts = append(opts, producthunt.WithTrendDateRange(from, to))
	} else if in.From != "" || in.To != "" {
		return nil, &mcptool.Error{
			Code:    "invalid_input",
			Message: "from and to must be provided together",
		}
	}
	if len(in.Topics) > 0 {
		opts = append(opts, producthunt.WithTrendTopics(in.Topics))
	}
	if in.TopN > 0 {
		opts = append(opts, producthunt.WithTrendTopN(in.TopN))
	}
	if len(in.StopWords) > 0 {
		opts = append(opts, producthunt.WithTrendStopWords(in.StopWords))
	}
	return c.AnalyzeTrends(ctx, opts...)
}

var trendsTools = []mcptool.Tool{
	mcptool.Define[*producthunt.Client, AnalyzeTrendsInput](
		"producthunt_analyze_trends",
		"Aggregate analytics over a Product Hunt date window: top products, keywords, topics, makers",
		"AnalyzeTrends",
		analyzeTrends,
	),
}
