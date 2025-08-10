package checks

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/abdultolba/nizam/internal/doctor"
)

type DockerDaemon struct{}

func (c DockerDaemon) ID() string { return "docker.daemon" }

func (c DockerDaemon) Run(ctx context.Context) (doctor.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return doctor.Result{ID: c.ID(), Status: doctor.Fail, Severity: "required", Message: "cannot init docker client"}, nil
	}
	defer cli.Close()
	
	if _, err := cli.Ping(ctx); err != nil {
		return doctor.Result{ID: c.ID(), Status: doctor.Fail, Severity: "required", Message: "docker daemon unreachable. Start Docker Desktop or dockerd"}, nil
	}
	sv, err := cli.ServerVersion(ctx)
	if err != nil {
		return doctor.Result{ID: c.ID(), Status: doctor.Warn, Severity: "advisory", Message: "cannot read docker version"}, nil
	}
	details := map[string]string{"version": sv.Version, "api": sv.APIVersion}
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "required", Details: details}, nil
}

func (c DockerDaemon) Fix(context.Context) error { return errors.New("no automatic fix") }

type ComposePlugin struct{}

func (c ComposePlugin) ID() string { return "docker.compose" }

func (c ComposePlugin) Run(ctx context.Context) (doctor.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "compose", "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return doctor.Result{
			ID: c.ID(), Status: doctor.Fail, Severity: "required",
			Message: "docker compose plugin not found",
			Hints: []string{"Install Docker Compose v2 plugin (bundled with Docker Desktop)"},
		}, nil
	}
	ver := strings.TrimSpace(string(out))
	return doctor.Result{ID: c.ID(), Status: doctor.OK, Severity: "required", Details: map[string]string{"version": ver}}, nil
}

func (c ComposePlugin) Fix(context.Context) error { return errors.New("install Docker Compose v2") }
