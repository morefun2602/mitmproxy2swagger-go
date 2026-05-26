package curate

import "testing"

func TestValidateProposedMatchesGroup(t *testing.T) {
	entries := []templateEntry{
		{ignore: true, path: "/items/foo"},
		{ignore: true, path: "/items/bar"},
	}
	if !validateProposedMatchesGroup("ignore:/items/{itemKey}", entries) {
		t.Fatal("expected valid match")
	}
	if validateProposedMatchesGroup("ignore:/v1/apps/info/{appKey}", entries) {
		t.Fatal("expected mismatch for wrong prefix")
	}
}
