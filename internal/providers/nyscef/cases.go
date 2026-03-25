package nyscef

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// tableRowRe matches a <tr> element and its contents.
var tableRowRe = regexp.MustCompile(`(?i)<tr[^>]*>([\s\S]*?)</tr>`)

// tableCellRe matches <td> element content (strips inner tags).
var tableCellRe = regexp.MustCompile(`(?i)<td[^>]*>([\s\S]*?)</td>`)

// htmlTagRe strips HTML tags for plain text extraction.
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// hrefRe extracts an href attribute value.
var hrefRe = regexp.MustCompile(`(?i)href="([^"]*)"`)

// docketIDRe extracts a docketId parameter from a URL.
var docketIDRe = regexp.MustCompile(`(?i)[?&]docketId=([^&"]+)`)

// indexNumberRe matches common index number patterns like "600001/2024".
var indexNumberRe = regexp.MustCompile(`\d{6}/\d{4}`)

func newCasesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cases",
		Short:   "Search and retrieve NYSCEF court cases",
		Aliases: []string{"case"},
	}

	cmd.AddCommand(newCasesSearchCmd(factory))
	cmd.AddCommand(newCasesGetCmd(factory))

	return cmd
}

// newCasesSearchCmd returns the `cases search` subcommand.
func newCasesSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search NYSCEF cases by county, type, or party name",
		RunE:  makeRunCasesSearch(factory),
	}
	cmd.Flags().String("county", "", "County name or code (bronx, kings/brooklyn, manhattan/new york, queens, richmond/staten island)")
	cmd.Flags().String("case-type", "", "Filter by case type keyword (e.g. probate, partition, estate)")
	cmd.Flags().String("party-name", "", "Search by party name")
	cmd.Flags().String("since", "", "Filter cases filed on or after this date (MM/DD/YYYY)")
	cmd.MarkFlagRequired("county")
	return cmd
}

func makeRunCasesSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		county, _ := cmd.Flags().GetString("county")
		caseType, _ := cmd.Flags().GetString("case-type")
		partyName, _ := cmd.Flags().GetString("party-name")
		since, _ := cmd.Flags().GetString("since")

		countyNum, ok := resolveCountyCode(county)
		if !ok {
			return fmt.Errorf("unknown county %q; use: bronx, kings/brooklyn, manhattan/new york, queens, richmond/staten island", county)
		}

		form := url.Values{}
		form.Set("selCounty", countyNum)
		form.Set("txtIndexNumber", "")
		form.Set("txtPartyName", partyName)
		form.Set("txtAttorneyName", "")
		form.Set("txtJudgeName", "")
		form.Set("txtFilingDateFrom", since)
		form.Set("txtFilingDateTo", "")
		form.Set("chkCaseType", "")
		form.Set("btnSubmit", "Search")

		searchURL := client.baseURL + "/nyscef/CaseSearch"
		body, err := client.Post(ctx, searchURL, form)
		if err != nil {
			return fmt.Errorf("search cases: %w", err)
		}

		cases, err := parseCaseSearchResults(body, client.baseURL)
		if err != nil {
			return fmt.Errorf("parse results: %w", err)
		}

		// Filter by case type keyword if provided.
		if caseType != "" {
			cases = filterByCaseType(cases, caseType)
		}

		return printCaseSummaries(cmd, cases)
	}
}

// newCasesGetCmd returns the `cases get` subcommand.
func newCasesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a NYSCEF case by docket ID",
		RunE:  makeRunCasesGet(factory),
	}
	cmd.Flags().String("docket-id", "", "NYSCEF docket ID")
	cmd.MarkFlagRequired("docket-id")
	return cmd
}

func makeRunCasesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		docketID, _ := cmd.Flags().GetString("docket-id")
		detailURL := client.baseURL + "/nyscef/CaseDetails?docketId=" + url.QueryEscape(docketID)

		body, err := client.Get(ctx, detailURL)
		if err != nil {
			return fmt.Errorf("get case: %w", err)
		}

		c, err := parseCaseDetail(body, docketID, detailURL)
		if err != nil {
			return fmt.Errorf("parse case: %w", err)
		}

		return printCaseDetail(cmd, c)
	}
}

// parseCaseSearchResults extracts CaseSummary entries from the NYSCEF CaseSearch HTML response.
// The results page renders a table where each row is a case.
func parseCaseSearchResults(body []byte, baseURL string) ([]CaseSummary, error) {
	html := string(body)

	// Check for "no results" indicators.
	if strings.Contains(html, "No cases found") || strings.Contains(html, "no results") {
		return []CaseSummary{}, nil
	}

	rows := tableRowRe.FindAllStringSubmatch(html, -1)
	if len(rows) == 0 {
		return []CaseSummary{}, nil
	}

	var cases []CaseSummary
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		rowContent := row[1]

		// Skip header rows (contain <th> tags).
		if strings.Contains(strings.ToLower(rowContent), "<th") {
			continue
		}

		cells := tableCellRe.FindAllStringSubmatch(rowContent, -1)
		if len(cells) < 1 {
			continue
		}

		// Extract plain text from each cell.
		cellTexts := make([]string, len(cells))
		for i, cell := range cells {
			if len(cell) >= 2 {
				cellTexts[i] = cleanHTML(cell[1])
			}
		}

		// Extract docket ID from any href in the row.
		docketID := extractDocketID(rowContent)
		caseURL := ""
		if docketID != "" {
			caseURL = baseURL + "/nyscef/CaseDetails?docketId=" + url.QueryEscape(docketID)
		}

		// Extract index number — look for pattern like "600001/2024" in cells.
		indexNumber := extractIndexNumber(cellTexts)

		// Build summary from available cells.
		// Typical column order: Index Number, Case Type, Caption, Filed Date, Court, Status
		c := CaseSummary{
			DocketID: docketID,
			URL:      caseURL,
		}

		switch len(cellTexts) {
		case 1:
			c.IndexNumber = cellTexts[0]
		case 2:
			c.IndexNumber = cellTexts[0]
			c.Caption = cellTexts[1]
		case 3:
			c.IndexNumber = cellTexts[0]
			c.Caption = cellTexts[1]
			c.FilingDate = cellTexts[2]
		case 4:
			c.IndexNumber = cellTexts[0]
			c.CaseType = cellTexts[1]
			c.Caption = cellTexts[2]
			c.FilingDate = cellTexts[3]
		case 5:
			c.IndexNumber = cellTexts[0]
			c.CaseType = cellTexts[1]
			c.Caption = cellTexts[2]
			c.FilingDate = cellTexts[3]
			c.Court = cellTexts[4]
		default:
			// 6+ columns: Index, CaseType, Caption, FilingDate, Court, Status
			c.IndexNumber = cellTexts[0]
			c.CaseType = cellTexts[1]
			c.Caption = cellTexts[2]
			c.FilingDate = cellTexts[3]
			c.Court = cellTexts[4]
			c.Status = cellTexts[5]
		}

		// Override IndexNumber if regex found a cleaner match.
		if indexNumber != "" && c.IndexNumber == "" {
			c.IndexNumber = indexNumber
		}

		// Skip rows that look like they have no meaningful content.
		if c.IndexNumber == "" && c.Caption == "" && c.DocketID == "" {
			continue
		}

		cases = append(cases, c)
	}

	return cases, nil
}

// parseCaseDetail extracts a CaseSummary from a NYSCEF CaseDetails HTML page.
// It first scans labeled table rows, then falls back to h1/h2 for the caption.
func parseCaseDetail(body []byte, docketID, pageURL string) (CaseSummary, error) {
	html := string(body)

	c := CaseSummary{
		DocketID: docketID,
		URL:      pageURL,
	}

	// Scan table rows for labeled fields — handles any page structure.
	rows := tableRowRe.FindAllStringSubmatch(html, -1)
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		cells := tableCellRe.FindAllStringSubmatch(row[1], -1)
		if len(cells) < 2 {
			continue
		}
		label := strings.ToLower(cleanHTML(cells[0][1]))
		value := cleanHTML(cells[1][1])
		if value == "" {
			continue
		}
		switch {
		case strings.Contains(label, "index"):
			if c.IndexNumber == "" {
				c.IndexNumber = value
			}
		case strings.Contains(label, "caption"):
			if c.Caption == "" {
				c.Caption = value
			}
		case strings.Contains(label, "type"):
			if c.CaseType == "" {
				c.CaseType = value
			}
		case strings.Contains(label, "filed") || strings.Contains(label, "date"):
			if c.FilingDate == "" {
				c.FilingDate = value
			}
		case strings.Contains(label, "court"):
			if c.Court == "" {
				c.Court = value
			}
		case strings.Contains(label, "status"):
			if c.Status == "" {
				c.Status = value
			}
		}
	}

	// Fall back to h1/h2 for the caption if not found in table rows.
	if c.Caption == "" {
		c.Caption = extractTagContent(html, "h1")
	}
	if c.Caption == "" {
		c.Caption = extractTagContent(html, "h2")
	}

	return c, nil
}

// filterByCaseType keeps only cases where caption or caseType contains the keyword.
func filterByCaseType(cases []CaseSummary, keyword string) []CaseSummary {
	lower := strings.ToLower(keyword)
	var filtered []CaseSummary
	for _, c := range cases {
		if strings.Contains(strings.ToLower(c.Caption), lower) ||
			strings.Contains(strings.ToLower(c.CaseType), lower) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// cleanHTML strips HTML tags and normalizes whitespace from a string.
func cleanHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	// Collapse whitespace.
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// extractDocketID finds a docketId parameter in any href in the row HTML.
func extractDocketID(html string) string {
	// First try direct docketId parameter in href.
	m := docketIDRe.FindStringSubmatch(html)
	if len(m) >= 2 {
		v, err := url.QueryUnescape(m[1])
		if err != nil {
			return m[1]
		}
		return v
	}

	// Try href that contains CaseDetails.
	hrefs := hrefRe.FindAllStringSubmatch(html, -1)
	for _, h := range hrefs {
		if len(h) < 2 {
			continue
		}
		href := h[1]
		if strings.Contains(strings.ToLower(href), "casedetails") {
			m2 := docketIDRe.FindStringSubmatch(href)
			if len(m2) >= 2 {
				v, err := url.QueryUnescape(m2[1])
				if err != nil {
					return m2[1]
				}
				return v
			}
		}
	}
	return ""
}

// extractIndexNumber finds the first index number pattern in a slice of cell texts.
func extractIndexNumber(cellTexts []string) string {
	for _, t := range cellTexts {
		m := indexNumberRe.FindString(t)
		if m != "" {
			return m
		}
	}
	return ""
}

// extractTagContent returns the text content of the first occurrence of tagName in html.
func extractTagContent(html, tagName string) string {
	re := regexp.MustCompile(`(?i)<` + tagName + `[^>]*>([\s\S]*?)</` + tagName + `>`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return cleanHTML(m[1])
}
