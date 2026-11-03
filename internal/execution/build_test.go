package execution

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBuildProgramWithoutCompilation(t *testing.T) {
	duration, err := buildProgram(context.Background(), nil, time.Second)
	if duration != nil || err != nil {
		t.Fatalf("skipped build = %v, %v", duration, err)
	}
}

func TestBuildProgramRetainsTimingForEveryOutcome(t *testing.T) {
	for _, test := range []struct {
		name, mode, errorText string
	}{
		{"success", "output", ""},
		{"failure", "invalid-mode", "failed"},
		{"timeout", "sleep", "timed out"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.name == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}
			command := helperCommand(test.mode)
			duration, err := buildProgram(ctx, &command, time.Second)
			if duration == nil || *duration <= 0 {
				t.Fatalf("missing build timing: %v", duration)
			}
			if test.errorText == "" && err != nil || test.errorText != "" && (err == nil || !strings.Contains(err.Error(), test.errorText)) {
				t.Fatalf("build error = %v, want %q", err, test.errorText)
			}
		})
	}
}
