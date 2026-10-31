package catalog

import (
	"encoding/json"
	"testing"
)

func TestValidNodeID(t *testing.T) {
	if !ValidNodeID("hello-world") || !ValidNodeID("go") {
		t.Fatal("expected kebab-case IDs to be valid")
	}
	for _, id := range []string{"", "Hello", "-lead", "trail-", "double--dash", "has_underscore"} {
		if ValidNodeID(id) {
			t.Errorf("ValidNodeID(%q) = true", id)
		}
	}
}

func TestDifficultyValidate(t *testing.T) {
	for _, difficulty := range []Difficulty{DifficultyEasy, DifficultyMedium, DifficultyHard} {
		if err := difficulty.Validate(); err != nil {
			t.Errorf("Validate(%q): %v", difficulty, err)
		}
	}
	if err := Difficulty("Impossible").Validate(); err == nil {
		t.Fatal("unsupported difficulty returned nil error")
	}
}

func TestParentInfoJSONRoundTrip(t *testing.T) {
	want := ParentInfo{
		ID: "fundamentals", Title: "Fundamentals",
		Items: []ItemRef{{ID: "hello-world", Order: 1}},
	}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got ParentInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.ID != want.ID || len(got.Items) != 1 || got.Items[0] != want.Items[0] {
		t.Fatalf("metadata = %+v, want %+v", got, want)
	}
}
