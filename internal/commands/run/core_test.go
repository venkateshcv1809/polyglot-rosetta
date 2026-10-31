package runcmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	runapi "polyglot-rosetta/internal/api/run"
	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/domain/evaluation"
	"polyglot-rosetta/internal/domain/testcase"
)

type fakeHost struct {
	concept                           catalog.Concept
	resolveErr, ensureErr, executeErr error
	mutate                            func(*runapi.Result)
	executed                          int
}

func (h *fakeHost) Resolve(_ context.Context, id string) (catalog.Concept, error) {
	if h.resolveErr != nil {
		return catalog.Concept{}, h.resolveErr
	}
	full, _ := catalog.ConceptID(h.concept.Path)
	return (catalog.Catalog{ConceptsByID: map[string]catalog.Concept{full: h.concept}}).ResolveConcept(id)
}
func (h *fakeHost) EnsureAvailable(context.Context, string) error { return h.ensureErr }
func (h *fakeHost) Execute(_ context.Context, request runapi.Request, tests []testcase.TestCase) (runapi.Result, error) {
	h.executed++
	if h.executeErr != nil {
		return runapi.Result{}, h.executeErr
	}
	r := runapi.NewResult()
	r.ID, r.Language, r.Mode = request.ID, request.Language, request.Mode
	r.Status, r.TotalPassed = evaluation.StatusPassed, len(tests)
	for _, tc := range tests {
		r.TestResults = append(r.TestResults, runapi.TestResult{ID: tc.ID, Status: evaluation.TestPassed, Passed: true})
	}
	if h.mutate != nil {
		h.mutate(&r)
	}
	return r, nil
}
func sharedRunFixture() (*fakeHost, runapi.Request) {
	id, _ := catalog.ConceptID("fundamentals/hello-world")
	host := &fakeHost{concept: catalog.Concept{Path: "fundamentals/hello-world", TestCases: []testcase.TestCase{{ID: "public"}, {ID: "hidden", Hidden: true}}}}
	return host, runapi.Request{ID: id[:10], Language: " TypeScript ", Source: "source"}
}
func TestRunSharedSemantics(t *testing.T) {
	for _, mode := range []evaluation.Mode{evaluation.ModePublic, evaluation.ModeAll} {
		host, request := sharedRunFixture()
		request.Mode = mode
		result, err := Run(context.Background(), request, host)
		if err != nil || result.Status != evaluation.StatusPassed {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		full, _ := catalog.ConceptID(host.concept.Path)
		want := 1
		if mode == evaluation.ModeAll {
			want = 2
		}
		if result.ID != full || result.Language != "typescript" || result.TotalPassed != want || result.Concept != host.concept.Path {
			t.Fatalf("result=%+v", result)
		}
		if mode == evaluation.ModeAll && !result.TestResults[1].Hidden {
			t.Fatal("host controlled hidden metadata")
		}
	}
	host, request := sharedRunFixture()
	result, err := Run(context.Background(), request, host)
	if err != nil || result.Mode != evaluation.ModePublic {
		t.Fatalf("default mode: %+v %v", result, err)
	}
}
func TestRunStopsOnHostAndRequestErrors(t *testing.T) {
	sentinel := errors.New("host failure")
	for _, scenario := range []string{"resolve", "ensure", "execute", "missing", "source", "mode", "language", "empty suite", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			host, request := sharedRunFixture()
			ctx := context.Background()
			switch scenario {
			case "resolve":
				host.resolveErr = sentinel
			case "ensure":
				host.ensureErr = sentinel
			case "execute":
				host.executeErr = sentinel
			case "missing":
				request.ID = "c-00000000"
			case "source":
				request.Source = " "
			case "mode":
				request.Mode = "other"
			case "language":
				request.Language = "ruby"
			case "empty suite":
				host.concept.TestCases = nil
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			result, err := Run(ctx, request, host)
			if err == nil || result.Status != evaluation.StatusError || result.Error == "" {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			if scenario != "execute" && host.executed != 0 {
				t.Fatal("executed despite failed preparation")
			}
			if (scenario == "resolve" || scenario == "ensure" || scenario == "execute") && !errors.Is(err, sentinel) {
				t.Fatal("lost host error identity")
			}
		})
	}
}
func TestRunValidatesExecutionResults(t *testing.T) {
	for _, mutate := range []func(*runapi.Result){
		func(r *runapi.Result) { r.SchemaVersion = 2 }, func(r *runapi.Result) { r.ID = "wrong" },
		func(r *runapi.Result) { r.TotalPassed = 0 }, func(r *runapi.Result) { r.DurationMs = -1 },
		func(r *runapi.Result) { r.TestResults = nil }, func(r *runapi.Result) { r.TestResults[0].Passed = false },
		func(r *runapi.Result) { r.Status = evaluation.StatusTimeout },
	} {
		host, request := sharedRunFixture()
		host.mutate = mutate
		result, err := Run(context.Background(), request, host)
		if err == nil || !strings.Contains(result.Error, "invalid execution result") {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
	for _, status := range []evaluation.Status{evaluation.StatusFailed, evaluation.StatusTimeout, evaluation.StatusError} {
		host, request := sharedRunFixture()
		host.mutate = func(r *runapi.Result) {
			r.Status, r.TotalPassed, r.TotalFailed = status, 0, 1
			r.TestResults[0].Passed = false
			switch status {
			case evaluation.StatusFailed:
				r.TestResults[0].Status = evaluation.TestWrongAnswer
			case evaluation.StatusTimeout:
				r.TestResults[0].Status = evaluation.TestTimeout
			case evaluation.StatusError:
				r.TestResults[0].Status = evaluation.TestCrashed
			}
		}
		result, err := Run(context.Background(), request, host)
		if err != nil || result.Status != status {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
}
