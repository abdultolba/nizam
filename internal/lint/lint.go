package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/abdultolba/nizam/internal/config"
)

type Finding struct {
	Rule       string
	Path       string
	Message    string
	Suggestion string
	Severity   string // "warn" | "error"
}

type Report struct {
	Findings []Finding
}

func (r *Report) Add(f Finding) {
	r.Findings = append(r.Findings, f)
}

func NoLatest(cfg *config.Config) (rep Report) {
	for name, s := range cfg.Services {
		if s.Image == "" || !strings.Contains(s.Image, ":") || strings.HasSuffix(s.Image, ":latest") {
			rep.Add(Finding{
				Rule:       "no-latest",
				Path:       "services." + name + ".image",
				Severity:   "error",
				Message:    "image tag missing or ':latest' not allowed",
				Suggestion: "pin to a specific tag, e.g. 'redis:7.2'",
			})
		}
	}
	return
}

func PortsShape(cfg *config.Config) (rep Report) {
	re := regexp.MustCompile(`^[0-9]{2,5}:[0-9]{2,5}$`)
	for name, s := range cfg.Services {
		for i, p := range s.Ports {
			if !re.MatchString(p) {
				rep.Add(Finding{
					Rule:       "ports-shape",
					Path:       fmt.Sprintf("services.%s.ports[%d]", name, i),
					Severity:   "error",
					Message:    "port mapping must be 'host:container' (no ranges)",
					Suggestion: "use single mapping like '8080:80'",
				})
			}
		}
	}
	return
}

func LimitsRecommended(cfg *config.Config) (rep Report) {
	// Example policy: encourage resource limits (not shown in structs)
	// Treat as warning for now.
	for name := range cfg.Services {
		rep.Add(Finding{
			Rule:       "limits",
			Path:       "services." + name,
			Severity:   "warn",
			Message:    "consider setting CPU/memory limits for repeatable dev envs",
			Suggestion: "add 'resources: { cpus: \"1.0\", memory: \"512m\" }'",
		})
	}
	return
}

func Combine(reps ...Report) Report {
	out := Report{}
	for _, r := range reps {
		out.Findings = append(out.Findings, r.Findings...)
	}
	return out
}
