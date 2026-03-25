package trends

import (
	"context"
	"fmt"
	"strings"

	"github.com/groovili/gogtrends"
)

// TrendsService abstracts gogtrends for testability.
type TrendsService interface {
	InterestOverTime(ctx context.Context, keyword, geo, timeRange string) ([]TimePoint, error)
	Compare(ctx context.Context, keywords []string, geo, timeRange string) ([]CompareResult, error)
}

// ServiceFactory is the function signature for creating a TrendsService.
type ServiceFactory func(ctx context.Context) (TrendsService, error)

// realService wraps gogtrends package-level functions.
type realService struct{}

// DefaultServiceFactory returns a ServiceFactory that uses the real gogtrends library.
func DefaultServiceFactory() ServiceFactory {
	return func(ctx context.Context) (TrendsService, error) {
		return &realService{}, nil
	}
}

// InterestOverTime fetches interest-over-time data for a single keyword.
func (s *realService) InterestOverTime(ctx context.Context, keyword, geo, timeRange string) ([]TimePoint, error) {
	widgets, err := gogtrends.Explore(ctx, &gogtrends.ExploreRequest{
		ComparisonItems: []*gogtrends.ComparisonItem{{
			Keyword: keyword,
			Geo:     geo,
			Time:    timeRange,
		}},
		Category: 0,
		Property: "",
	}, "EN")
	if err != nil {
		return nil, fmt.Errorf("explore request: %w", err)
	}

	// Find the InterestOverTime widget (ID prefix "TIMESERIES").
	var timeWidget *gogtrends.ExploreWidget
	for _, w := range widgets {
		if strings.HasPrefix(w.ID, string(gogtrends.IntOverTimeWidgetID)) {
			timeWidget = w
			break
		}
	}
	if timeWidget == nil {
		return nil, fmt.Errorf("no interest-over-time widget returned for %q", keyword)
	}

	timeline, err := gogtrends.InterestOverTime(ctx, timeWidget, "EN")
	if err != nil {
		return nil, fmt.Errorf("interest over time: %w", err)
	}

	points := make([]TimePoint, 0, len(timeline))
	for _, t := range timeline {
		value := 0
		if len(t.Value) > 0 {
			value = t.Value[0]
		}
		points = append(points, TimePoint{
			Date:  t.FormattedTime,
			Value: value,
		})
	}
	return points, nil
}

// Compare fetches interest-over-time data for multiple keywords simultaneously.
func (s *realService) Compare(ctx context.Context, keywords []string, geo, timeRange string) ([]CompareResult, error) {
	if len(keywords) == 0 {
		return nil, fmt.Errorf("at least one keyword is required")
	}

	items := make([]*gogtrends.ComparisonItem, 0, len(keywords))
	for _, kw := range keywords {
		items = append(items, &gogtrends.ComparisonItem{
			Keyword: strings.TrimSpace(kw),
			Geo:     geo,
			Time:    timeRange,
		})
	}

	widgets, err := gogtrends.Explore(ctx, &gogtrends.ExploreRequest{
		ComparisonItems: items,
		Category:        0,
		Property:        "",
	}, "EN")
	if err != nil {
		return nil, fmt.Errorf("explore request: %w", err)
	}

	// Find the InterestOverTime widget.
	var timeWidget *gogtrends.ExploreWidget
	for _, w := range widgets {
		if strings.HasPrefix(w.ID, string(gogtrends.IntOverTimeWidgetID)) {
			timeWidget = w
			break
		}
	}
	if timeWidget == nil {
		return nil, fmt.Errorf("no interest-over-time widget returned")
	}

	timeline, err := gogtrends.InterestOverTime(ctx, timeWidget, "EN")
	if err != nil {
		return nil, fmt.Errorf("interest over time: %w", err)
	}

	// Build per-keyword result arrays. Each Timeline.Value has one entry per keyword.
	results := make([]CompareResult, len(keywords))
	for i, kw := range keywords {
		results[i] = CompareResult{
			Keyword: strings.TrimSpace(kw),
			Data:    make([]TimePoint, 0, len(timeline)),
		}
	}

	for _, t := range timeline {
		for i := range keywords {
			value := 0
			if i < len(t.Value) {
				value = t.Value[i]
			}
			results[i].Data = append(results[i].Data, TimePoint{
				Date:  t.FormattedTime,
				Value: value,
			})
		}
	}

	return results, nil
}
