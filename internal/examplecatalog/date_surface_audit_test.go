package examplecatalog

import (
	"path/filepath"
	"testing"
)

func TestDateSurfaceAuditRowsCoverUpstreamFamilies(t *testing.T) {
	want := []string{
		"dates.py:function:date2num",
		"dates.py:function:num2date",
		"dates.py:class:DateFormatter",
		"dates.py:class:AutoDateFormatter",
		"dates.py:class:ConciseDateFormatter",
		"dates.py:class:DateLocator",
		"dates.py:class:AutoDateLocator",
		"dates.py:class:RRuleLocator",
		"dates.py:class:YearLocator",
		"dates.py:class:MonthLocator",
		"dates.py:class:WeekdayLocator",
		"dates.py:class:DayLocator",
		"dates.py:class:HourLocator",
		"dates.py:class:MinuteLocator",
		"dates.py:class:SecondLocator",
		"dates.py:class:MicrosecondLocator",
		"dates.py:class:DateConverter",
		"dates.py:class:ConciseDateConverter",
		"dates.py:class:rrulewrapper",
	}

	rows := DateSurfaceAuditRows()
	seen := map[string]DateSurfaceAudit{}
	for _, row := range rows {
		if row.UpstreamID == "" || row.Status == "" || row.Note == "" || len(row.GoFiles) == 0 {
			t.Fatalf("incomplete date surface audit row: %+v", row)
		}
		if !validPublicSurfaceParityStatus(row.Status) {
			t.Fatalf("%s has invalid status %q", row.UpstreamID, row.Status)
		}
		if _, ok := seen[row.UpstreamID]; ok {
			t.Fatalf("duplicate date surface audit row %q", row.UpstreamID)
		}
		seen[row.UpstreamID] = row
	}
	for _, id := range want {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing date surface audit row %q", id)
		}
	}
}

func TestDateSurfaceAuditRowsReferenceExistingLocalFiles(t *testing.T) {
	root := repoRoot(t)
	for _, row := range DateSurfaceAuditRows() {
		for _, path := range row.GoFiles {
			requireFile(t, filepath.Join(root, path))
		}
	}
}
