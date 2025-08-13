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
	RequiredFailed int `json:"required_failed"`
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
	fmt.Println("🔍 Running system diagnostics...")
	fmt.Println()

	// Track stats for final summary
	var totalChecks, passedChecks, failedChecks, warnChecks int
	var systemReady = true
	var recommendations []string

	for _, c := range r.Checks {
		totalChecks++
		
		// Skip user-skipped checks in main output
		if c.Status == Warn && c.Message == "skipped by user" {
			continue
		}

		icon := "✓"
		colorCode := "\033[32m" // green
		resetCode := "\033[0m"
		
		if c.Status == Fail {
			icon = "✗"
			colorCode = "\033[31m" // red
			failedChecks++
			systemReady = false
		} else if c.Status == Warn {
			icon = "⚠"
			colorCode = "\033[33m" // yellow
			warnChecks++
			if c.Severity == "required" {
				systemReady = false
			}
		} else {
			passedChecks++
		}

		// Generate human-readable description
		description := getCheckDescription(c)
		
		fmt.Printf("%s%s %s%s\n", colorCode, icon, description, resetCode)
		
		// Show error message if failed
		if c.Status == Fail && c.Message != "" {
			fmt.Printf("  \033[31m✗ %s\033[0m\n", c.Message)
		}
		
		// Show warning details for MTU and other warnings
		if c.Status == Warn && c.Message != "" && c.Message != "skipped by user" {
			fmt.Printf("  \033[33m⚠ %s\033[0m\n", c.Message)
		}

		// Collect recommendations from hints
		for _, hint := range c.Hints {
			if c.Status == Fail {
				// Show immediate fixes for failed checks
				fmt.Printf("  \033[36m• %s\033[0m\n", hint)
			} else if c.Status == Warn {
				// Collect recommendations for warnings
				recommendations = append(recommendations, hint)
			}
		}
	}

	fmt.Println()

	// Final status
	if systemReady {
		fmt.Println("🎉 System ready for Nizam!")
	} else {
		fmt.Println("❌ System needs attention before running Nizam")
	}

	// Show recommendations if any
	if len(recommendations) > 0 {
		fmt.Println("\nRecommendations:")
		for _, rec := range recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}

	// Show summary if there are failures or advisory issues
	if r.Summary.RequiredFailed > 0 || r.Summary.AdvisoryFailed > 0 {
		fmt.Printf("\nIssues found: %d required, %d advisory\n", r.Summary.RequiredFailed, r.Summary.AdvisoryFailed)
	}

	// Count skipped checks
	skippedCount := 0
	for _, c := range r.Checks {
		if c.Status == Warn && c.Message == "skipped by user" {
			skippedCount++
		}
	}
	if skippedCount > 0 {
		fmt.Printf("\nSkipped %d check(s) by user request\n", skippedCount)
	}
}

func getCheckDescription(c Result) string {
	// Extract version info from details if available
	var versionInfo string
	if details, ok := c.Details.(map[string]interface{}); ok {
		if version, exists := details["version"]; exists {
			versionInfo = fmt.Sprintf(" %v", version)
		}
		if freeBytes, exists := details["free_bytes"]; exists {
			if bytes, ok := freeBytes.(uint64); ok {
				gb := float64(bytes) / (1024 * 1024 * 1024)
				versionInfo = fmt.Sprintf(": %.1f GB", gb)
			}
		}
	}
	
	switch c.ID {
	case "docker.daemon":
		return fmt.Sprintf("Docker daemon connectivity%s", versionInfo)
	case "docker.compose":
		return "Docker Compose plugin available"
	case "disk.free":
		return fmt.Sprintf("Available disk space%s", versionInfo)
	case "memory.usage":
		if details, ok := c.Details.(map[string]interface{}); ok {
			if usedGB, exists := details["used_gb"]; exists {
				if totalGB, exists := details["total_gb"]; exists {
					if percent, exists := details["percent"]; exists {
						return fmt.Sprintf("Memory usage: %s/%s GB (%d%%)", usedGB, totalGB, percent)
					}
				}
			}
		}
		return "Memory usage"
	case "net.mtu":
		return "Network MTU configuration"
	default:
		// Handle port checks
		if strings.HasPrefix(c.ID, "port.") {
			port := strings.TrimPrefix(c.ID, "port.")
			return fmt.Sprintf("Port %s availability", port)
		}
		return c.ID
	}
}

func (r Report) PrintJSON() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(r)
}

var ErrRequiredFailed = errors.New("required checks failed")
