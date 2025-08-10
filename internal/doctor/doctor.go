package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
)

type Status string

const (
	OK   Status = "ok"
	Fail Status = "fail"
	Warn Status = "warn"
)

type Result struct {
	ID       string      `json:"id"`
	Status   Status      `json:"status"`
	Severity string      `json:"severity,omitempty"` // "required" | "advisory"
	Message  string      `json:"message,omitempty"`
	Hints    []string    `json:"hints,omitempty"`
	Details  interface{} `json:"details,omitempty"`
}

type Check interface {
	ID() string
	Run(ctx context.Context) (Result, error)
	Fix(ctx context.Context) error // no-op if unsupported
}

type Runner struct {
	Checks      []Check
	MaxParallel int
	Timeout     time.Duration
}

type Summary struct {
	RequiredFailed  int `json:"required_failed"`
	AdvisoryFailed int `json:"advisory_failed"`
}

type Report struct {
	Summary Summary  `json:"summary"`
	Checks  []Result `json:"checks"`
}

func (r *Runner) Run(ctx context.Context, skip map[string]struct{}, doFix bool) (Report, error) {
	if r.MaxParallel <= 0 {
		r.MaxParallel = 6
	}
	if r.Timeout <= 0 {
		r.Timeout = 8 * time.Second
	}

	sem := semaphore.NewWeighted(int64(r.MaxParallel))
	results := make([]Result, len(r.Checks))

	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	doneCh := make(chan struct{}, len(r.Checks))

	for i, c := range r.Checks {
		i, c := i, c
		if _, skipThis := skip[c.ID()]; skipThis {
			results[i] = Result{ID: c.ID(), Status: Warn, Severity: "advisory", Message: "skipped by user"}
			doneCh <- struct{}{}
			continue
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			return Report{}, err
		}
		go func() {
			defer sem.Release(1)
			res, err := c.Run(ctx)
			if err != nil {
				// If check returns an error, record as fail with message
				res = Result{ID: c.ID(), Status: Fail, Severity: "required", Message: err.Error()}
			}
			results[i] = res
			if doFix && res.Status != OK {
				_ = c.Fix(ctx) // best-effort
			}
			doneCh <- struct{}{}
		}()
	}

	for range r.Checks {
		select {
		case <-doneCh:
		case <-ctx.Done():
			return Report{}, ctx.Err()
		}
	}

	// Sort by severity then id for stable output
	sort.Slice(results, func(i, j int) bool {
		if results[i].Severity == results[j].Severity {
			return results[i].ID < results[j].ID
		}
		return results[i].Severity < results[j].Severity
	})

	rep := Report{Checks: results}
	for _, re := range results {
		switch strings.ToLower(re.Severity) {
		case "required":
			if re.Status == Fail {
				rep.Summary.RequiredFailed++
			}
		case "advisory":
			if re.Status != OK {
				rep.Summary.AdvisoryFailed++
			}
		}
	}
	return rep, nil
}

func (r Report) PrintHuman() {
	for _, c := range r.Checks {
		icon := "✔"
		if c.Status == Fail {
			icon = "✖"
		} else if c.Status == Warn {
			icon = "!"
		}
		line := fmt.Sprintf("%s %-20s", icon, c.ID)
		if c.Message != "" {
			line += " " + c.Message
		}
		fmt.Println(line)
		for _, h := range c.Hints {
			fmt.Printf("  %s\n", h)
		}
	}
	fmt.Printf("\nSummary: required_failed=%d advisory_failed=%d\n", r.Summary.RequiredFailed, r.Summary.AdvisoryFailed)
}

func (r Report) PrintJSON() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(r)
}

var ErrRequiredFailed = errors.New("required checks failed")
