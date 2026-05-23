package examplecatalog

import "testing"

func TestCoverageAuditExitCriteriaCapturePlanCriteria(t *testing.T) {
	want := []string{
		"feature-area-classification",
		"implemented-public-feature-fixtures",
		"user-facing-demo-breadth",
		"browser-demo-catalog-alignment",
	}
	for _, id := range want {
		if _, ok := LookupCoverageAuditExitCriterion(id); !ok {
			t.Fatalf("missing coverage audit exit criterion %q", id)
		}
	}
}

func TestCoverageAuditExitCriteriaAreActionable(t *testing.T) {
	for _, criterion := range CoverageAuditExitCriteria() {
		if criterion.ID == "" || criterion.Criterion == "" || criterion.Interpretation == "" {
			t.Fatalf("incomplete coverage audit exit criterion: %+v", criterion)
		}
		switch criterion.Status {
		case AuditCriterionSatisfied:
		case AuditCriterionFollowUpRequired:
			if criterion.FollowUp == "" {
				t.Fatalf("%s requires follow-up but has no follow-up text", criterion.ID)
			}
		default:
			t.Fatalf("%s has unsupported status %q", criterion.ID, criterion.Status)
		}
		if len(criterion.Evidence) == 0 {
			t.Fatalf("%s has no evidence references", criterion.ID)
		}
	}
}

func TestCoverageAuditExitCriteriaExposeRemainingWork(t *testing.T) {
	var followUp int
	for _, criterion := range CoverageAuditExitCriteria() {
		if criterion.Status == AuditCriterionFollowUpRequired {
			followUp++
		}
	}
	if followUp == 0 {
		t.Fatal("coverage audit exit criteria should expose remaining implementation/demo work")
	}
}
