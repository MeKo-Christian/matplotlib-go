package examplecatalog

import (
	"fmt"
	"sort"
	"strings"
)

type markdownTableAlign int

const (
	markdownTableLeft markdownTableAlign = iota
	markdownTableRight
)

// MatplotlibParityStatusMarkdown renders the committed parity inventories as a
// compact status ledger for docs/matplotlib-parity-status.md.
func MatplotlibParityStatusMarkdown(publicSurface []PublicSurfaceRow) string {
	publicRows := PublicSurfaceParityRowsForSurface(publicSurface)
	sort.Slice(publicRows, func(i, j int) bool {
		if publicRows[i].FeatureCoverageID != publicRows[j].FeatureCoverageID {
			return publicRows[i].FeatureCoverageID < publicRows[j].FeatureCoverageID
		}
		return publicRows[i].UpstreamID < publicRows[j].UpstreamID
	})

	var b strings.Builder
	b.WriteString("# Matplotlib Parity Status\n\n")
	b.WriteString("Generated from `internal/examplecatalog` and `test/testdata/parity_surface/upstream_public_surface.json`.\n\n")
	b.WriteString("## Feature Coverage\n\n")
	featureRows := make([][]string, 0, len(FeatureCoverageMatrix()))
	for _, row := range FeatureCoverageMatrix() {
		featureRows = append(featureRows, []string{
			markdownEscape(row.ID),
			string(row.GoEquivalent),
			string(row.ParityFixture),
			string(row.UserShowcase),
			string(row.BrowserDemo),
			string(row.Breadth),
		})
	}
	writeMarkdownTable(
		&b,
		[]string{"Feature", "Go", "Fixture", "Showcase", "Browser", "Breadth"},
		[]markdownTableAlign{markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft},
		featureRows,
	)

	b.WriteString("\n## Browser Demo Coverage\n\n")
	browserRows := make([][]string, 0, len(BrowserDemoCoverageRows()))
	for _, row := range BrowserDemoCoverageRows() {
		reference := row.ReferenceModule
		if reference == "" {
			reference = "-"
		}
		demoID := row.ActiveWebDemoID
		if demoID == "" {
			demoID = "-"
		}
		browserRows = append(browserRows, []string{
			markdownEscape(row.ID),
			string(row.Status),
			markdownEscape(demoID),
			markdownEscape(strings.Join(row.CatalogIDs, ", ")),
			markdownEscape(reference),
			markdownEscape(row.Action),
		})
	}
	writeMarkdownTable(
		&b,
		[]string{"Row", "Status", "Browser demo", "Catalog cases", "Reference", "Action"},
		[]markdownTableAlign{markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft},
		browserRows,
	)

	b.WriteString("\n## Public Surface Summary\n\n")
	writeStatusSummary(&b, publicRows)
	b.WriteString("\n")
	writeClosureSummary(&b, publicRows)

	b.WriteString("\n## Open Public Surface Rows\n\n")
	openRows := make([][]string, 0, len(publicRows))
	for _, row := range publicRows {
		if row.Status != PublicSurfacePartial && row.Status != PublicSurfaceNotStarted {
			continue
		}
		owner := row.ClosurePhase
		if owner == "" {
			owner = row.ClosureRationale
		}
		openRows = append(openRows, []string{
			markdownEscapeUpstreamID(row.UpstreamID),
			markdownEscape(row.FeatureCoverageID),
			string(row.Status),
			markdownEscape(owner),
			markdownEscape(strings.Join(row.GoFiles, ", ")),
			markdownEscape(row.Note),
		})
	}
	writeMarkdownTable(
		&b,
		[]string{"Upstream", "Feature", "Status", "Closure owner", "Local files", "Note"},
		[]markdownTableAlign{markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft},
		openRows,
	)

	b.WriteString("\n## Per-Family Status\n\n")
	b.WriteString("One table per upstream feature family. Columns: upstream API or registry item, Go status, local API file(s), parity fixture, user example, browser demo, remaining work.\n\n")
	writePerFamilyTables(&b, publicRows)
	return b.String()
}

func writePerFamilyTables(b *strings.Builder, rows []PublicSurfaceParity) {
	// Group rows by feature family.
	familyRows := make(map[string][]PublicSurfaceParity)
	for _, row := range rows {
		familyRows[row.FeatureCoverageID] = append(familyRows[row.FeatureCoverageID], row)
	}
	families := make([]string, 0, len(familyRows))
	for f := range familyRows {
		families = append(families, f)
	}
	sort.Strings(families)

	// Build lookup: catalogID → active demo IDs.
	catalogToDemos := make(map[string][]string)
	for _, demo := range BrowserDemoCoverageRows() {
		if demo.Status != BrowserDemoActive {
			continue
		}
		demoID := demo.ActiveWebDemoID
		if demoID == "" {
			continue
		}
		for _, catalogID := range demo.CatalogIDs {
			catalogToDemos[catalogID] = append(catalogToDemos[catalogID], demoID)
		}
	}

	for _, family := range families {
		frows := familyRows[family]
		sort.Slice(frows, func(i, j int) bool { return frows[i].UpstreamID < frows[j].UpstreamID })

		b.WriteString("### " + family + "\n\n")

		tableRows := make([][]string, 0, len(frows))
		for _, row := range frows {
			localAPI := "-"
			if len(row.GoFiles) > 0 {
				localAPI = strings.Join(row.GoFiles, ", ")
			}

			fixture := strings.Join(row.CatalogIDs, ", ")
			if fixture == "" {
				fixture = "-"
			}

			example := strings.Join(row.ExampleIDs, ", ")
			if example == "" {
				example = "-"
			}

			// Collect unique active browser demos for this row's catalog cases.
			demoSet := make(map[string]bool)
			for _, catalogID := range row.CatalogIDs {
				for _, demo := range catalogToDemos[catalogID] {
					demoSet[demo] = true
				}
			}
			demoList := make([]string, 0, len(demoSet))
			for demo := range demoSet {
				demoList = append(demoList, demo)
			}
			sort.Strings(demoList)
			browserDemo := strings.Join(demoList, ", ")
			if browserDemo == "" {
				browserDemo = "-"
			}

			remaining := "-"
			switch row.Status {
			case PublicSurfacePartial, PublicSurfaceNotStarted:
				remaining = row.Note
			case PublicSurfaceIntentionalOmission:
				remaining = row.Note
				if remaining == "" {
					remaining = "intentional omission"
				}
			}

			tableRows = append(tableRows, []string{
				markdownEscapeUpstreamID(row.UpstreamID),
				string(row.Status),
				markdownEscape(localAPI),
				markdownEscape(fixture),
				markdownEscape(example),
				markdownEscape(browserDemo),
				markdownEscape(remaining),
			})
		}

		writeMarkdownTable(
			b,
			[]string{"Upstream", "Status", "Local API", "Fixtures", "Examples", "Browser demo", "Remaining work"},
			[]markdownTableAlign{
				markdownTableLeft, markdownTableLeft, markdownTableLeft,
				markdownTableLeft, markdownTableLeft, markdownTableLeft, markdownTableLeft,
			},
			tableRows,
		)
		b.WriteString("\n")
	}
}

func writeStatusSummary(b *strings.Builder, rows []PublicSurfaceParity) {
	counts := map[PublicSurfaceParityStatus]int{}
	for _, row := range rows {
		counts[row.Status]++
	}
	tableRows := make([][]string, 0, 5)
	for _, status := range []PublicSurfaceParityStatus{
		PublicSurfaceDirectEquivalent,
		PublicSurfaceIdiomaticEquivalent,
		PublicSurfacePartial,
		PublicSurfaceNotStarted,
		PublicSurfaceIntentionalOmission,
	} {
		tableRows = append(tableRows, []string{string(status), fmt.Sprintf("%d", counts[status])})
	}
	writeMarkdownTable(
		b,
		[]string{"Status", "Rows"},
		[]markdownTableAlign{markdownTableLeft, markdownTableRight},
		tableRows,
	)
}

func writeClosureSummary(b *strings.Builder, rows []PublicSurfaceParity) {
	counts := map[string]int{}
	for _, row := range rows {
		if row.Status != PublicSurfacePartial && row.Status != PublicSurfaceNotStarted {
			continue
		}
		owner := row.ClosurePhase
		if owner == "" {
			owner = row.ClosureRationale
		}
		counts[owner]++
	}
	owners := make([]string, 0, len(counts))
	for owner := range counts {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	b.WriteString("## Closure Owner Summary\n\n")
	tableRows := make([][]string, 0, len(owners))
	for _, owner := range owners {
		tableRows = append(tableRows, []string{markdownEscape(owner), fmt.Sprintf("%d", counts[owner])})
	}
	writeMarkdownTable(
		b,
		[]string{"Owner", "Open rows"},
		[]markdownTableAlign{markdownTableLeft, markdownTableRight},
		tableRows,
	)
}

func markdownEscape(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "*", "\\*")
	value = markdownEscapeLeadingUnderscores(value)
	return value
}

func markdownEscapeUpstreamID(value string) string {
	value = markdownEscape(value)
	var b strings.Builder
	b.Grow(len(value))
	for i, r := range value {
		if r == '_' && (i == 0 || value[i-1] == '/' || value[i-1] == ':') {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func markdownEscapeLeadingUnderscores(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for i, r := range value {
		if r == '_' && markdownUnderscoreCanOpenEmphasis(value, i) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func markdownUnderscoreCanOpenEmphasis(value string, idx int) bool {
	if idx+1 >= len(value) || !markdownASCIIAlnum(value[idx+1]) {
		return false
	}
	if idx == 0 {
		return true
	}
	return !markdownASCIIAlnum(value[idx-1]) && value[idx-1] != '\\'
}

func markdownASCIIAlnum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func writeMarkdownTable(b *strings.Builder, headers []string, aligns []markdownTableAlign, rows [][]string) {
	if len(headers) != len(aligns) {
		panic("markdown table headers and alignments differ")
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		if len(row) != len(headers) {
			panic("markdown table row has unexpected width")
		}
		for i, cell := range row {
			widths[i] = max(widths[i], len(cell))
		}
	}

	writeMarkdownTableRow(b, headers, aligns, widths)
	separators := make([]string, len(headers))
	for i, width := range widths {
		if width < 3 {
			width = 3
		}
		switch aligns[i] {
		case markdownTableRight:
			separators[i] = strings.Repeat("-", width-1) + ":"
		default:
			separators[i] = strings.Repeat("-", width)
		}
	}
	writeMarkdownTableRow(b, separators, aligns, widths)
	for _, row := range rows {
		writeMarkdownTableRow(b, row, aligns, widths)
	}
}

func writeMarkdownTableRow(b *strings.Builder, cells []string, aligns []markdownTableAlign, widths []int) {
	b.WriteString("|")
	for i, cell := range cells {
		if aligns[i] == markdownTableRight {
			fmt.Fprintf(b, " %*s |", widths[i], cell)
			continue
		}
		fmt.Fprintf(b, " %-*s |", widths[i], cell)
	}
	b.WriteString("\n")
}
