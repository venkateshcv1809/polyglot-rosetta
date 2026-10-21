package scaffoldcmd

import "testing"

func TestParseCategoryOptions(t *testing.T) {
	opts, err := parseCategoryOptions("generate category", []string{
		"--parent=algorithms",
		"--id=sorting",
		"--title=Sorting",
		"--description=Ordering algorithms",
	})
	if err != nil {
		t.Fatalf("parseCategoryOptions: %v", err)
	}
	if opts.parent != "algorithms" || opts.id != "sorting" || opts.title != "Sorting" || opts.description != "Ordering algorithms" {
		t.Fatalf("options = %+v", opts)
	}
}

func TestParseCategoryOptionsRejectsPositionalArguments(t *testing.T) {
	if _, err := parseCategoryOptions("generate category", []string{"sorting"}); err == nil {
		t.Fatal("parseCategoryOptions returned nil error")
	}
}
