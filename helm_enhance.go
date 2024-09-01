package search

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

// Define constants to avoid duplication of literal strings
const (
	chartNina          = "niña"
	chartPinta         = "pinta"
	chartSantaMaria    = "santa-maria"
	chartPintaVersion2 = "2.0.0"
	chartSantaMariaVer = "1.2.3"
	chartSantaMariaRC  = "1.2.2-RC-1"
	repoTesting        = "testing"
	repoZTesting       = "ztesting"
)

// TestSortScore verifies the correctness of sorting by score and version.
func TestSortScore(t *testing.T) {
	in := []*Result{
		{Name: "bbb", Score: 0, Chart: &repo.ChartVersion{Metadata: &chart.Metadata{Version: "1.2.3"}}},
		{Name: "aaa", Score: 5},
		{Name: "abb", Score: 5},
		{Name: "aab", Score: 0},
		{Name: "bab", Score: 5},
		{Name: "ver", Score: 5, Chart: &repo.ChartVersion{Metadata: &chart.Metadata{Version: "1.2.4"}}},
		{Name: "ver", Score: 5, Chart: &repo.ChartVersion{Metadata: &chart.Metadata{Version: "1.2.3"}}},
	}
	expect := []string{"aab", "bbb", "aaa", "abb", "bab", "ver", "ver"}
	expectScore := []int{0, 0, 5, 5, 5, 5, 5}
	SortScore(in)

	// Test Score
	for i, expectedScore := range expectScore {
		if expectedScore != in[i].Score {
			t.Errorf("Sort error on index %d: expected %d, got %d", i, expectedScore, in[i].Score)
		}
	}
	// Test Name
	for i, expectedName := range expect {
		if expectedName != in[i].Name {
			t.Errorf("Sort error: expected %s, got %s", expectedName, in[i].Name)
		}
	}

	// Test version of the last two items
	if in[5].Chart.Metadata.Version != "1.2.4" {
		t.Errorf("Expected 1.2.4, got %s", in[5].Chart.Metadata.Version)
	}
	if in[6].Chart.Metadata.Version != "1.2.3" {
		t.Error("Expected 1.2.3 to be last")
	}
}

// indexfileEntries contains predefined chart entries.
var indexfileEntries = map[string]repo.ChartVersions{
	chartNina: {
		{
			URLs: []string{"http://example.com/charts/nina-0.1.0.tgz"},
			Metadata: &chart.Metadata{
				Name:        chartNina,
				Version:     "0.1.0",
				Description: "One boat",
			},
		},
	},
	chartPinta: {
		{
			URLs: []string{"http://example.com/charts/pinta-0.1.0.tgz"},
			Metadata: &chart.Metadata{
				Name:        chartPinta,
				Version:     "0.1.0",
				Description: "Two ship",
			},
		},
	},
	chartSantaMaria: {
		{
			URLs: []string{"http://example.com/charts/santa-maria-1.2.3.tgz"},
			Metadata: &chart.Metadata{
				Name:        chartSantaMaria,
				Version:     "1.2.3",
				Description: "Three boat",
			},
		},
		{
			URLs: []string{"http://example.com/charts/santa-maria-1.2.2-rc-1.tgz"},
			Metadata: &chart.Metadata{
				Name:        chartSantaMaria,
				Version:     "1.2.2-RC-1",
				Description: "Three boat",
			},
		},
	},
}

// loadTestIndex initializes a new Index with predefined chart entries.
func loadTestIndex(_ *testing.T, all bool) *Index {
	i := NewIndex()
	i.AddRepo(repoTesting, &repo.IndexFile{Entries: indexfileEntries}, all)
	i.AddRepo(repoZTesting, &repo.IndexFile{Entries: map[string]repo.ChartVersions{
		chartPinta: {
			{
				URLs: []string{"http://example.com/charts/pinta-2.0.0.tgz"},
				Metadata: &chart.Metadata{
					Name:        chartPinta,
					Version:     chartPintaVersion2,
					Description: "Two ship, version two",
				},
			},
		},
	}}, all)
	return i
}

// TestRepoEntries verifies the number of entries based on the "all" flag.
func TestRepoEntries(t *testing.T) {
	i := loadTestIndex(t, false)
	all := i.All()
	if len(all) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(all))
	}

	i = loadTestIndex(t, true)
	all = i.All()
	if len(all) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(all))
	}
}

// TestSortRepoEntries verifies sorting of charts by version after adding a repo.
func TestSortRepoEntries(t *testing.T) {
	i := loadTestIndex(t, true)
	sr, err := i.Search("TESTING/SANTA-MARIA", 100, false)
	if err != nil {
		t.Fatal(err)
	}
	SortScore(sr)

	ch := sr[0]
	expect := chartSantaMariaVer
	if ch.Chart.Metadata.Version != expect {
		t.Errorf("Expected %q, got %q", expect, ch.Chart.Metadata.Version)
	}
}

// TestSearchByName tests the search functionality by name and description.
func TestSearchByName(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		expect  []*Result
		regexp  bool
		fail    bool
		failMsg string
	}{
		{
			name:  "basic search for one result",
			query: chartSantaMaria,
			expect: []*Result{
				{Name: repoTesting + "/" + chartSantaMaria},
			},
		},
		{
			name:  "basic search for two results",
			query: chartPinta,
			expect: []*Result{
				{Name: repoTesting + "/" + chartPinta},
				{Name: repoZTesting + "/Pinta"},
			},
		},
		{
			name:  "repo-specific search for one result",
			query: repoZTesting + "/" + chartPinta,
			expect: []*Result{
				{Name: repoZTesting + "/Pinta"},
			},
		},
		{
			name:  "partial name search",
			query: "santa",
			expect: []*Result{
				{Name: repoTesting + "/" + chartSantaMaria},
			},
		},
		{
			name:  "description search, one result",
			query: "Three",
			expect: []*Result{
				{Name: repoTesting + "/" + chartSantaMaria},
			},
		},
		{
			name:  "description search, two results",
			query: "two",
			expect: []*Result{
				{Name: repoTesting + "/" + chartPinta},
				{Name: repoZTesting + "/Pinta"},
			},
		},
		{
			name:  "search mixedCase and result should be mixedCase too",
			query: chartPinta,
			expect: []*Result{
				{Name: repoTesting + "/" + chartPinta},
				{Name: repoZTesting + "/Pinta"},
			},
		},
		{
			name:  "description upper search, two results",
			query: "TWO",
			expect: []*Result{
				{Name: repoTesting + "/" + chartPinta},
				{Name: repoZTesting + "/Pinta"},
			},
		},
		{
			name:   "nothing found",
			query:  "mayflower",
			expect: []*Result{},
		},
		{
			name:    "regexp, one result",
			query:   "Th[ref]*",
			expect:  []*Result{{Name: repoTesting + "/" + chartSantaMaria}},
			regexp:  true,
		},
		{
			name:    "regexp, fail compile",
			query:   "th[",
			expect:  []*Result{},
			regexp:  true,
			fail:    true,
			failMsg: "error parsing regexp:",
		},
	}

	i := loadTestIndex(t, false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charts, err := i.Search(tt.query, 100, tt.regexp)
			if err != nil {
				if tt.fail {
					if !strings.Contains(err.Error(), tt.failMsg) {
						t.Fatalf("Unexpected error message: %s", err)
					}
					return
				}
				t.Fatalf("%s: %s", tt.name, err)
			}
			if len(charts) != len(tt.expect) {
				t.Errorf("Expected %d results, got %d", len(tt.expect), len(charts))
			}
			for i, chart := range charts {
				if chart.Name != tt.expect[i].Name {
					t.Errorf("Expected %s, got %s", tt.expect[i].Name, chart.Name)
				}
			}
		})
	}
}

// TestSearchByNameAll tests search functionality with the "all" flag set.
func TestSearchByNameAll(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		expect  int
		regexp  bool
	}{
		{name: "one result", query: chartSantaMaria, expect: 1},
		{name: "two results", query: chartPinta, expect: 2},
		{name: "partial name search", query: "santa", expect: 1},
		{name: "description search, one result", query: "Three", expect: 1},
		{name: "description search, two results", query: "two", expect: 2},
		{name: "search mixedCase", query: chartPinta, expect: 2},
		{name: "description upper search", query: "TWO", expect: 2},
		{name: "nothing found", query: "mayflower", expect: 0},
		{name: "regexp", query: "Th[ref]*", expect: 1, regexp: true},
	}

	i := loadTestIndex(t, true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charts, err := i.Search(tt.query, 100, tt.regexp)
			if err != nil {
				t.Fatalf("%s: %s", tt.name, err)
			}
			if len(charts) != tt.expect {
				t.Errorf("Expected %d results, got %d", tt.expect, len(charts))
			}
		})
	}
}
