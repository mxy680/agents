package trends

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// --- interest search ---

func TestInterestSearch_TextOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Jan 2025") {
		t.Errorf("expected date in output, got: %s", out)
	}
	if !strings.Contains(out, "DATE") {
		t.Errorf("expected header in output, got: %s", out)
	}
}

func TestInterestSearch_JSONOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TimePoint
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) != 12 {
		t.Errorf("expected 12 data points, got %d", len(results))
	}
	if results[0].Date == "" {
		t.Error("expected non-empty date")
	}
	if results[0].Value == 0 {
		t.Error("expected non-zero value")
	}
}

func TestInterestSearch_CustomGeoAndTime(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"trends", "interest", "search",
			"--keyword=bed stuy", "--geo=US", "--time=today 3-m", "--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []TimePoint
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if len(results) == 0 {
		t.Error("expected at least one data point")
	}
}

func TestInterestSearch_ServiceError(t *testing.T) {
	svc := &mockService{err: errors.New("trends API unavailable")}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when service fails")
	}
}

func TestInterestSearch_MissingKeyword(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "search"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --keyword is missing")
	}
}

// --- interest compare ---

func TestInterestCompare_TextOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"trends", "interest", "compare",
			"--keywords=mott haven,east new york,bed stuy",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "mott haven") {
		t.Errorf("expected keyword in output, got: %s", out)
	}
}

func TestInterestCompare_JSONOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"trends", "interest", "compare",
			"--keywords=mott haven,east new york", "--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []CompareResult
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 compare results, got %d", len(results))
	}
	if results[0].Keyword != "mott haven" {
		t.Errorf("expected keyword 'mott haven', got %q", results[0].Keyword)
	}
	if len(results[0].Data) == 0 {
		t.Error("expected non-empty data for first keyword")
	}
}

func TestInterestCompare_ServiceError(t *testing.T) {
	svc := &mockService{err: errors.New("trends API unavailable")}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{
		"trends", "interest", "compare",
		"--keywords=mott haven,east new york",
	})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when service fails")
	}
}

func TestInterestCompare_MissingKeywords(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "compare"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --keywords is missing")
	}
}

func TestInterestCompare_EmptyKeywordsAfterSplit(t *testing.T) {
	// --keywords flag present but contains only commas/whitespace — splitKeywords returns empty.
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "compare", "--keywords=,,"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when keywords are all empty after splitting")
	}
}

// --- interest momentum ---

func TestInterestMomentum_TextOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "momentum", "--keyword=mott haven"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "mott haven") {
		t.Errorf("expected keyword in output, got: %s", out)
	}
	if !strings.Contains(out, "rising") {
		t.Errorf("expected 'rising' trend in output, got: %s", out)
	}
}

func TestInterestMomentum_JSONOutput(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "momentum", "--keyword=mott haven", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result MomentumResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if result.Keyword != "mott haven" {
		t.Errorf("expected keyword 'mott haven', got %q", result.Keyword)
	}
	if result.Trend != "rising" {
		t.Errorf("expected trend 'rising', got %q", result.Trend)
	}
	if result.MomentumPct <= 15 {
		t.Errorf("expected momentum > 15%%, got %.2f", result.MomentumPct)
	}
}

func TestInterestMomentum_ServiceError(t *testing.T) {
	svc := &mockService{err: errors.New("trends API unavailable")}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "momentum", "--keyword=mott haven"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when service fails")
	}
}

func TestInterestMomentum_MissingKeyword(t *testing.T) {
	svc := &mockService{}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "momentum"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --keyword is missing")
	}
}

// --- calculateMomentum unit tests ---

func TestCalculateMomentum_Rising(t *testing.T) {
	points := buildDefaultTimePoints()
	result, err := calculateMomentum("test keyword", points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Trend != "rising" {
		t.Errorf("expected 'rising', got %q", result.Trend)
	}
	if result.Keyword != "test keyword" {
		t.Errorf("expected keyword 'test keyword', got %q", result.Keyword)
	}
	// first 3 avg = (30+32+28)/3 = 30; last 3 avg = (55+60+65)/3 = 60
	if result.EarlierAvg != 30 {
		t.Errorf("expected earlier avg 30, got %.2f", result.EarlierAvg)
	}
	if result.RecentAvg != 60 {
		t.Errorf("expected recent avg 60, got %.2f", result.RecentAvg)
	}
	if result.MomentumPct != 100 {
		t.Errorf("expected momentum 100%%, got %.2f", result.MomentumPct)
	}
}

func TestCalculateMomentum_Declining(t *testing.T) {
	// Reverse of the default dataset — clearly declining.
	rawValues := []int{65, 60, 55, 50, 45, 42, 38, 40, 35, 28, 32, 30}
	points := make([]TimePoint, len(rawValues))
	for i, v := range rawValues {
		points[i] = TimePoint{Date: defaultTestDates[i], Value: v}
	}

	result, err := calculateMomentum("declining kw", points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Trend != "declining" {
		t.Errorf("expected 'declining', got %q", result.Trend)
	}
	if result.MomentumPct >= -15 {
		t.Errorf("expected momentum < -15%%, got %.2f", result.MomentumPct)
	}
}

func TestCalculateMomentum_Stable(t *testing.T) {
	// Flat data — no momentum in either direction.
	points := make([]TimePoint, 12)
	for i := range points {
		points[i] = TimePoint{Date: defaultTestDates[i], Value: 50}
	}

	result, err := calculateMomentum("stable kw", points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Trend != "stable" {
		t.Errorf("expected 'stable', got %q", result.Trend)
	}
	if result.MomentumPct != 0 {
		t.Errorf("expected 0%% momentum for flat data, got %.2f", result.MomentumPct)
	}
}

func TestCalculateMomentum_InsufficientData(t *testing.T) {
	points := buildDefaultTimePoints()[:4] // only 4 points
	_, err := calculateMomentum("kw", points)
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestCalculateMomentum_EarlierAvgZero(t *testing.T) {
	// Earlier three months all zero — should not divide by zero.
	points := make([]TimePoint, 12)
	for i := range points {
		v := 0
		if i >= 9 {
			v = 50
		}
		points[i] = TimePoint{Date: defaultTestDates[i], Value: v}
	}

	result, err := calculateMomentum("kw", points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MomentumPct != 0 {
		t.Errorf("expected 0%% momentum when earlier avg is zero, got %.2f", result.MomentumPct)
	}
}

// --- factory construction error paths ---

func TestInterestSearch_FactoryConstructionError(t *testing.T) {
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newErrServiceFactory()}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when factory fails")
	}
}

func TestInterestCompare_FactoryConstructionError(t *testing.T) {
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newErrServiceFactory()}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "compare", "--keywords=mott haven"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when factory fails")
	}
}

func TestInterestMomentum_FactoryConstructionError(t *testing.T) {
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newErrServiceFactory()}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "momentum", "--keyword=mott haven"})
	if err := root.Execute(); err == nil {
		t.Error("expected error when factory fails")
	}
}

// --- empty data paths ---

func TestInterestSearch_EmptyData(t *testing.T) {
	svc := &mockService{interestData: []TimePoint{}}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No data found") {
		t.Errorf("expected 'No data found' message, got: %s", out)
	}
}

func TestInterestSearch_EmptyDataJSON(t *testing.T) {
	svc := &mockService{interestData: []TimePoint{}}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "search", "--keyword=mott haven", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	var results []TimePoint
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty array, got %d results", len(results))
	}
}

func TestInterestCompare_EmptyData(t *testing.T) {
	svc := &mockService{compareData: []CompareResult{}}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "interest", "compare", "--keywords=mott haven"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No data found") {
		t.Errorf("expected 'No data found' message, got: %s", out)
	}
}

func TestInterestMomentum_InsufficientData(t *testing.T) {
	// Only 4 data points — momentum calculation should fail.
	svc := &mockService{interestData: buildDefaultTimePoints()[:4]}
	root := newTestRootCmd()
	p := &Provider{ServiceFactory: newMockServiceFactory(svc)}
	p.RegisterCommands(root)

	root.SetArgs([]string{"trends", "interest", "momentum", "--keyword=mott haven"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

// --- avgValues unit tests ---

func TestAvgValues_Empty(t *testing.T) {
	if got := avgValues(nil); got != 0 {
		t.Errorf("avgValues(nil) = %.2f, want 0", got)
	}
	if got := avgValues([]TimePoint{}); got != 0 {
		t.Errorf("avgValues([]) = %.2f, want 0", got)
	}
}

// --- DefaultServiceFactory ---

func TestDefaultServiceFactory(t *testing.T) {
	factory := DefaultServiceFactory()
	if factory == nil {
		t.Fatal("DefaultServiceFactory() returned nil")
	}
	// The factory itself must be callable and return a non-nil service.
	// We do not make real network calls.
	svc, err := factory(nil)
	if err != nil {
		t.Fatalf("factory returned error: %v", err)
	}
	if svc == nil {
		t.Fatal("factory returned nil service")
	}
}

// --- splitKeywords unit tests ---

func TestSplitKeywords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"mott haven,east new york,bed stuy", []string{"mott haven", "east new york", "bed stuy"}},
		{"  mott haven , east new york ", []string{"mott haven", "east new york"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{",,,", []string{}},
	}
	for _, tt := range tests {
		got := splitKeywords(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitKeywords(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("splitKeywords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

// --- classifyTrend unit tests ---

func TestClassifyTrend(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{100, "rising"},
		{16, "rising"},
		{15, "stable"},
		{0, "stable"},
		{-15, "stable"},
		{-16, "declining"},
		{-100, "declining"},
	}
	for _, tt := range tests {
		got := classifyTrend(tt.pct)
		if got != tt.want {
			t.Errorf("classifyTrend(%.1f) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}
