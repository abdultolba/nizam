package checks

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/doctor"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
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
				fmt.Sprintf("Change host port for service %s in .nizam.yaml", c.Svc),
				"Or stop the process using the port",
			},
		}, nil
	}
	_ = ln.Close()
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "required"}, nil
}

func (c PortInUse) Fix(ctx context.Context) error { return nil }

type MemoryUsage struct{}

func (c MemoryUsage) ID() string { return "memory.usage" }

func (c MemoryUsage) Run(ctx context.Context) (doctor.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return doctor.Result{ID: c.ID(), Status: doctor.Warn, Severity: "advisory", Message: "cannot determine memory usage"}, nil
	}
	
	usedGB := float64(v.Used) / (1024 * 1024 * 1024)
	totalGB := float64(v.Total) / (1024 * 1024 * 1024)
	usagePercent := int(v.UsedPercent)
	
	details := map[string]interface{}{
		"used_gb":    fmt.Sprintf("%.1f", usedGB),
		"total_gb":   fmt.Sprintf("%.1f", totalGB),
		"percent":    usagePercent,
		"available": fmt.Sprintf("%.1f", float64(v.Available)/(1024*1024*1024)),
	}
	
	// Warn if memory usage is very high
	if v.UsedPercent > 90 {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Warn, Severity: "advisory",
			Message: fmt.Sprintf("High memory usage: %.1f/%.1f GB (%d%%)", usedGB, totalGB, usagePercent),
			Details: details,
			Hints: []string{"Consider closing other applications to free up memory"},
		}, nil
	} else if totalGB < 4.0 {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Warn, Severity: "advisory",
			Message: fmt.Sprintf("Low system memory: %.1f GB total", totalGB),
			Details: details,
			Hints: []string{"Consider upgrading RAM for better Docker performance"},
		}, nil
	}
	
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "advisory", Details: details}, nil
}

func (c MemoryUsage) Fix(context.Context) error { return nil }
