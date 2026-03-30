package trends

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// mockService is a test double for TrendsService that returns predictable data.
type mockService struct {
	// interestData is returned by InterestOverTime. If nil, uses defaultInterestData.
	interestData []TimePoint
	// compareData is returned by Compare. If nil, derives from interestData.
	compareData []CompareResult
	// err is returned by all methods when non-nil.
	err error
}

// defaultTestValues is the canonical 12-month dataset used across tests.
// Values give a clear "rising" momentum signal: first 3 avg = 30, last 3 avg = 60.
var defaultTestValues = []int{30, 32, 28, 35, 40, 38, 42, 45, 50, 55, 60, 65}

// defaultTestDates are the formatted dates corresponding to defaultTestValues.
var defaultTestDates = []string{
	"Jan 2025", "Feb 2025", "Mar 2025", "Apr 2025",
	"May 2025", "Jun 2025", "Jul 2025", "Aug 2025",
	"Sep 2025", "Oct 2025", "Nov 2025", "Dec 2025",
}

// buildDefaultTimePoints returns a []TimePoint from defaultTestValues.
func buildDefaultTimePoints() []TimePoint {
	points := make([]TimePoint, len(defaultTestValues))
	for i, v := range defaultTestValues {
		points[i] = TimePoint{Date: defaultTestDates[i], Value: v}
	}
	return points
}

// InterestOverTime returns canned time point data.
func (m *mockService) InterestOverTime(_ context.Context, _, _, _ string) ([]TimePoint, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.interestData != nil {
		return m.interestData, nil
	}
	return buildDefaultTimePoints(), nil
}

// Compare returns canned compare result data.
func (m *mockService) Compare(_ context.Context, keywords []string, _, _ string) ([]CompareResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.compareData != nil {
		return m.compareData, nil
	}
	// Default: build a CompareResult per keyword using defaultTestValues.
	results := make([]CompareResult, len(keywords))
	points := buildDefaultTimePoints()
	for i, kw := range keywords {
		results[i] = CompareResult{Keyword: kw, Data: points}
	}
	return results, nil
}

// newMockServiceFactory returns a ServiceFactory that always returns the given mockService.
func newMockServiceFactory(svc *mockService) ServiceFactory {
	return func(_ context.Context) (TrendsService, error) {
		return svc, nil
	}
}

// newErrServiceFactory returns a ServiceFactory that always fails at construction time.
func newErrServiceFactory() ServiceFactory {
	return func(_ context.Context) (TrendsService, error) {
		return nil, errors.New("service construction error")
	}
}

// newTestRootCmd creates a root command with --json and --dry-run flags wired up.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
	root.PersistentFlags().Bool("dry-run", false, "Preview actions without executing them")
	return root
}

// captureStdout captures stdout during f() and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 256*1024)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}
