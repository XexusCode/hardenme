package check_test

import (
	"testing"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
)

func TestRelevanceDependsOnProfile(t *testing.T) {
	filevault := macCheck(t, "macos.filevault")
	lockdown := macCheck(t, "macos.lockdown_mode")
	gatekeeper := macCheck(t, "macos.gatekeeper")

	tests := []struct {
		name string
		c    check.Check
		p    profile.Profile
		want check.Relevance
	}{
		{"disk encryption on a laptop", filevault, profile.Profile{Portable: true, LeavesHome: true, Risk: profile.RiskNormal}, check.Essential},
		{"disk encryption on a desktop that stays home", filevault, profile.Profile{Risk: profile.RiskNormal}, check.Optional},
		{"disk encryption on a shared desktop", filevault, profile.Profile{Shared: true, Risk: profile.RiskNormal}, check.Recommended},
		{"disk encryption on a desktop with banking data", filevault, profile.Profile{SensitiveData: true, Risk: profile.RiskNormal}, check.Essential},
		{"lockdown mode at normal risk", lockdown, profile.Profile{Risk: profile.RiskNormal}, check.NotApplicable},
		{"lockdown mode at elevated risk", lockdown, profile.Profile{Risk: profile.RiskElevated}, check.Recommended},
		{"gatekeeper is always essential", gatekeeper, profile.Profile{Risk: profile.RiskNormal}, check.Essential},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.RelevanceFor(tc.p); got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestRelevanceTakesTheHighestMatchingLevel(t *testing.T) {
	yes := true
	c := check.Check{Relevance: check.RelevanceSpec{
		Default: check.Optional,
		Rules: []check.RelevanceRule{
			{When: check.Condition{Portable: &yes}, Level: check.Essential},
			{When: check.Condition{Shared: &yes}, Level: check.Recommended},
		},
	}}
	p := profile.Profile{Portable: true, Shared: true}
	if got := c.RelevanceFor(p); got != check.Essential {
		t.Fatalf("got %s, want essential", got)
	}
}
