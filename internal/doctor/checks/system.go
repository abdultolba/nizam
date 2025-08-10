package checks

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/abdultolba/nizam/internal/doctor"
)

type DiskFree struct {
	Path     string
	MinBytes uint64 // e.g., 5*GiB
}

func (c DiskFree) ID() string { return "disk.free" }

func (c DiskFree) Run(ctx context.Context) (doctor.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	usage, err := disk.UsageWithContext(ctx, c.Path)
	if err != nil {
		return doctor.Result{ID: c.ID(), Status: doctor.Warn, Severity: "advisory", Message: "cannot determine disk usage"}, nil
	}
	if usage.Free < c.MinBytes {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Fail, Severity: "required",
			Message: fmt.Sprintf("low disk space on %s: free=%dMiB", c.Path, usage.Free/1024/1024),
			Hints:   []string{"Prune with: docker system prune -af --volumes"},
		}, nil
	}
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "required", Details: map[string]any{"free_bytes": usage.Free}}, nil
}

func (c DiskFree) Fix(context.Context) error { return nil }

type MTUCheck struct{}

func (c MTUCheck) ID() string { return "net.mtu" }

func (c MTUCheck) Run(ctx context.Context) (doctor.Result, error) {
	// Advisory heuristic: warn if any interface MTU differs by >= 50 from 1500
	ifaces, _ := net.Interfaces()
	var msgs []string
	for _, inf := range ifaces {
		if inf.MTU <= 0 {
			continue
		}
		if abs(inf.MTU-1500) >= 50 {
			msgs = append(msgs, fmt.Sprintf("%s=%d", inf.Name, inf.MTU))
		}
	}
	if len(msgs) > 0 {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Warn, Severity: "advisory",
			Message: "non-standard MTU detected",
			Details: map[string]any{"ifaces": strings.Join(msgs, ", ")},
			Hints:   []string{"VPNs may lower MTU; if Docker networking is flaky, align MTU in daemon.json"},
		}, nil
	}
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "advisory"}, nil
}

func (c MTUCheck) Fix(context.Context) error { return nil }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type PortInUse struct {
	Host string
	Port int
	Svc  string
}

func (c PortInUse) ID() string { return fmt.Sprintf("port.%d", c.Port) }

func (c PortInUse) Run(ctx context.Context) (doctor.Result, error) {
	// Try to bind; if binding fails, port is in use.
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Fail, Severity: "required",
			Message: "port in use", Details: map[string]any{"addr": addr},
			Hints: []string{
				fmt.Sprintf("Change host port for service %s", c.Svc),
				"Or run: nizam up --resolve-ports",
			},
		}, nil
	}
	_ = ln.Close()
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "required"}, nil
}

func (c PortInUse) Fix(ctx context.Context) error { return nil }
