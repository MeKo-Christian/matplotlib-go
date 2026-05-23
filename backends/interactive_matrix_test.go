package backends_test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestInteractiveMatrixHasFigureFactoryForEveryCatalogTopic(t *testing.T) {
	for _, row := range examplecatalog.InteractiveCoverageMatrix() {
		if row.RepresentativeID == "" {
			t.Fatalf("topic %q has no representative case", row.Topic)
		}
		if _, _, err := parity.Figure(row.RepresentativeID); err != nil {
			t.Fatalf("topic %q representative %q is not figure-backed: %v", row.Topic, row.RepresentativeID, err)
		}
	}
}

func TestInteractiveMatrixCoversEveryCatalogTopic(t *testing.T) {
	topics := map[string]bool{}
	for _, c := range examplecatalog.Cases() {
		topics[c.Topic] = true
	}
	rows := examplecatalog.InteractiveCoverageMatrix()
	if len(rows) != len(topics) {
		t.Fatalf("matrix rows = %d, catalog topics = %d", len(rows), len(topics))
	}
	for _, row := range rows {
		if !topics[row.Topic] {
			t.Fatalf("matrix topic %q is not in catalog", row.Topic)
		}
		if !row.WebAgg || !row.Gio {
			t.Fatalf("topic %q must cover WebAgg and Gio, got webagg=%v gio=%v", row.Topic, row.WebAgg, row.Gio)
		}
		delete(topics, row.Topic)
	}
	for topic := range topics {
		t.Fatalf("missing interactive coverage row for topic %q", topic)
	}
}
