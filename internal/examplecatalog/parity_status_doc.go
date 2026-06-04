package examplecatalog

import (
	"fmt"
	"sort"
	"strings"
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
	b.WriteString("| Feature | Go | Fixture | Showcase | Browser | Breadth |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, row := range FeatureCoverageMatrix() {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			markdownEscape(row.ID),
			row.GoEquivalent,
			row.ParityFixture,
			row.UserShowcase,
			row.BrowserDemo,
			row.Breadth)
	}

	b.WriteString("\n## Public Surface Summary\n\n")
	writeStatusSummary(&b, publicRows)
	b.WriteString("\n")
	writeClosureSummary(&b, publicRows)

	b.WriteString("\n## Open Public Surface Rows\n\n")
	b.WriteString("| Upstream | Feature | Status | Closure owner | Local files | Note |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, row := range publicRows {
		if row.Status != PublicSurfacePartial && row.Status != PublicSurfaceNotStarted {
			continue
		}
		owner := row.ClosurePhase
		if owner == "" {
			owner = row.ClosureRationale
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			markdownEscape(row.UpstreamID),
			markdownEscape(row.FeatureCoverageID),
			row.Status,
			markdownEscape(owner),
			markdownEscape(strings.Join(row.GoFiles, ", ")),
			markdownEscape(row.Note))
	}
	return b.String()
}

func writeStatusSummary(b *strings.Builder, rows []PublicSurfaceParity) {
	counts := map[PublicSurfaceParityStatus]int{}
	for _, row := range rows {
		counts[row.Status]++
	}
	b.WriteString("| Status | Rows |\n")
	b.WriteString("| --- | ---: |\n")
	for _, status := range []PublicSurfaceParityStatus{
		PublicSurfaceDirectEquivalent,
		PublicSurfaceIdiomaticEquivalent,
		PublicSurfacePartial,
		PublicSurfaceNotStarted,
		PublicSurfaceIntentionalOmission,
	} {
		fmt.Fprintf(b, "| %s | %d |\n", status, counts[status])
	}
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
	b.WriteString("| Owner | Open rows |\n")
	b.WriteString("| --- | ---: |\n")
	for _, owner := range owners {
		fmt.Fprintf(b, "| %s | %d |\n", markdownEscape(owner), counts[owner])
	}
}

func markdownEscape(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
