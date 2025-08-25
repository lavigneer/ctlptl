package cluster

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/tilt-dev/ctlptl/internal/dctr"
	"github.com/tilt-dev/ctlptl/pkg/api"
)

// rancherMachine manages the Rancher Desktop virtual machine and Kubernetes cluster
// using the improved RancherDesktopManager that leverages rdctl's efficient APIs.
type rancherMachine struct {
	iostreams    genericclioptions.IOStreams
	dockerClient dctr.Client
	sleep        sleeper
	rdm          RancherManager
	os           string
}

func NewRancherMachine(ctx context.Context, client dctr.Client, iostreams genericclioptions.IOStreams) (*rancherMachine, error) {
	rdm, err := NewRancherDesktopManager()
	if err != nil {
		return nil, err
	}

	return &rancherMachine{
		dockerClient: client,
		iostreams:    iostreams,
		sleep:        time.Sleep,
		rdm:          rdm,
		os:           runtime.GOOS,
	}, nil
}

func (m *rancherMachine) CPUs(ctx context.Context) (int, error) {
	// Get CPU count from Docker info
	info, err := m.dockerClient.Info(ctx)
	if err != nil {
		return 0, err
	}
	return info.NCPU, nil
}

func (m *rancherMachine) EnsureExists(ctx context.Context) error {
	// Check if Kubernetes is running
	running, err := m.rdm.IsKubernetesRunning(ctx)
	if err != nil || !running {
		// Start with default settings if not running
		return m.rdm.Start(ctx, m.defaultCluster())
	}
	return nil
}

func (m *rancherMachine) defaultCluster() *api.Cluster {
	return &api.Cluster{
		MinCPUs:           2,
		KubernetesVersion: "1.31.11",
	}
}

func (m *rancherMachine) Restart(ctx context.Context, desired, existing *api.Cluster) error {
	// For Rancher Desktop, restart means reset and start again
	if err := m.rdm.ResetCluster(ctx); err != nil {
		return fmt.Errorf("failed to reset Rancher Desktop cluster: %v", err)
	}
	return m.rdm.Start(ctx, desired)
}

// Apply method is not part of the Machine interface, so we'll remove it
// The cluster controller handles the Apply logic
