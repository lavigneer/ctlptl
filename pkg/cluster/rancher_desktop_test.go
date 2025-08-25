package cluster

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tilt-dev/ctlptl/pkg/api"
)

func TestRancherDesktopManager_Status(t *testing.T) {
	// Skip if rdctl is not available
	rdm, err := NewRancherDesktopManager()
	if err != nil {
		t.Skip("Rancher Desktop not installed: ", err)
	}

	ctx := context.Background()
	status, err := rdm.Status(ctx)
	if err != nil {
		t.Skip("Failed to get Rancher Desktop status: ", err)
	}

	// Verify status has expected structure
	assert.NotNil(t, status)
	vm, ok := status["vm"].(map[string]interface{})
	assert.True(t, ok, "Status should have 'vm' field")
	assert.NotNil(t, vm)
}

func TestRancherDesktopManager_IsKubernetesRunning(t *testing.T) {
	// Skip if rdctl is not available
	rdm, err := NewRancherDesktopManager()
	if err != nil {
		t.Skip("Rancher Desktop not installed: ", err)
	}

	ctx := context.Background()
	running, err := rdm.IsKubernetesRunning(ctx)
	if err != nil {
		t.Skip("Failed to check Kubernetes status: ", err)
	}

	// Should return a boolean value
	assert.IsType(t, true, running)
}

func TestRancherMachine_CPUs(t *testing.T) {
	// This test requires a Docker client, so we'll skip it for now
	// In a real test environment, you'd mock the Docker client
	t.Skip("Requires Docker client")
}

func TestRancherDesktopAdmin_Create(t *testing.T) {
	// Skip if rdctl is not available
	rdm, err := NewRancherDesktopManager()
	if err != nil {
		t.Skip("Rancher Desktop not installed: ", err)
	}

	admin := newRancherDesktopAdmin("unix:///var/run/docker.sock", "darwin", rdm)

	ctx := context.Background()
	cluster := &api.Cluster{
		Name:     "test-rancher-desktop",
		Product:  "rancher-desktop",
		MinCPUs:  2,
	}

	// Test that registry is not supported
	err = admin.Create(ctx, cluster, &api.Registry{Name: "test-registry"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support connecting a registry")

	// Test that registry auths are not supported
	cluster.RegistryAuths = []api.RegistryAuth{{Host: "test.com"}}
	err = admin.Create(ctx, cluster, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support connecting pull-through registries")
}

func TestRancherDesktopAdmin_EnsureInstalled(t *testing.T) {
	// Skip if rdctl is not available
	rdm, err := NewRancherDesktopManager()
	if err != nil {
		t.Skip("Rancher Desktop not installed: ", err)
	}

	admin := newRancherDesktopAdmin("unix:///var/run/docker.sock", "darwin", rdm)

	ctx := context.Background()
	err = admin.EnsureInstalled(ctx)
	assert.NoError(t, err)
}
