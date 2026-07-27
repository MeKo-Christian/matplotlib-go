package examplecatalog

// AuditCriterionStatus records whether a coverage-audit exit criterion is
// satisfied by the audit inventory itself or still requires follow-up work.
type AuditCriterionStatus string

const (
	AuditCriterionSatisfied        AuditCriterionStatus = "satisfied"
	AuditCriterionFollowUpRequired AuditCriterionStatus = "follow-up-required"
)

// AuditExitCriterion records the current interpretation and evidence for one
// coverage-audit exit criterion.
type AuditExitCriterion struct {
	ID             string
	Criterion      string
	Interpretation string
	Status         AuditCriterionStatus
	Evidence       []string
	FollowUp       string
}

var coverageAuditExitCriteria = []AuditExitCriterion{
	{
		ID:             "feature-area-classification",
		Criterion:      "Every fundamental Matplotlib feature area is classified as implemented, partially implemented, intentionally omitted, or pending.",
		Interpretation: "The audit is complete when each tracked feature area has explicit status values; the status may still be partial or pending.",
		Status:         AuditCriterionSatisfied,
		Evidence: []string{
			"internal/examplecatalog.FeatureCoverageMatrix",
			"internal/examplecatalog.TestFeatureCoverageRowsHaveStableIDsAndStatuses",
			"internal/examplecatalog.TestFeatureCoverageCoversFundamentalMatplotlibAreas",
		},
	},
	{
		ID:             "implemented-public-feature-fixtures",
		Criterion:      "Every implemented public feature has at least one parity fixture or a documented reason why visual parity testing is not applicable.",
		Interpretation: "The audit is complete when implemented rows cannot exist without catalog coverage or an explicit omission note, and catalog cases have canonical Go/Python/reference sources.",
		Status:         AuditCriterionSatisfied,
		Evidence: []string{
			"internal/examplecatalog.TestImplementedFeatureCoverageHasFixtureOrOmission",
			"internal/examplecatalog.TestCatalogCasesHaveCanonicalParitySourcePairs",
			"internal/examplecatalog.TestCatalogCasesHaveVisibleMatplotlibReferenceModules",
		},
	},
	{
		ID:             "user-facing-demo-breadth",
		Criterion:      "Every major user-facing feature family has a showcase demo, and broad features have demos that exercise meaningful variants rather than only the minimal path.",
		Interpretation: "The audit identifies missing or thin demo breadth; it does not claim every demo has already been implemented.",
		Status:         AuditCriterionFollowUpRequired,
		Evidence: []string{
			"internal/examplecatalog.DemoBreadthGaps",
			"internal/examplecatalog.TestHighPriorityDemoBreadthGapsPointToThinCoverage",
		},
		FollowUp: "Promote high- and medium-priority DemoBreadthGaps into targeted demo implementation phases.",
	},
	{
		ID:             "browser-demo-catalog-alignment",
		Criterion:      "Browser demo coverage is aligned with the catalog instead of being a separate, drifting subset.",
		Interpretation: "The audit accounts for catalog and Python web reference drift, but many rows remain planned browser-demo wiring tasks.",
		Status:         AuditCriterionFollowUpRequired,
		Evidence: []string{
			"internal/examplecatalog.BrowserDemoCoverageRows",
			"internal/examplecatalog.TestBrowserDemoCoverageReferencesCatalogAndWebDemos",
			"internal/examplecatalog.TestBrowserCoverageAccountsForCLIOnlyShowcases",
		},
		FollowUp: "Wire planned BrowserDemoCoverageRows into active browser demos or explicitly keep them reference-only.",
	},
}

// CoverageAuditExitCriteria returns the current exit-criteria interpretation
// for the feature/demo coverage audit.
func CoverageAuditExitCriteria() []AuditExitCriterion {
	out := make([]AuditExitCriterion, len(coverageAuditExitCriteria))
	copy(out, coverageAuditExitCriteria)
	return out
}

// LookupCoverageAuditExitCriterion finds a coverage-audit exit criterion by
// stable ID.
func LookupCoverageAuditExitCriterion(id string) (AuditExitCriterion, bool) {
	for _, criterion := range CoverageAuditExitCriteria() {
		if criterion.ID == id {
			return criterion, true
		}
	}
	return AuditExitCriterion{}, false
}
