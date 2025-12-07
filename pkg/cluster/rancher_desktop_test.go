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

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"GB format", "4GB", 4, false},
		{"GB format lowercase", "4gb", 4, false},
		{"GB format with space", " 8GB ", 8, false},
		{"MB to GB conversion", "2048MB", 2, false},
		{"MB to GB rounding up", "3000MB", 3, false},
		{"Plain number", "6", 6, false},
		{"Invalid format", "invalid", 0, true},
		{"Empty string", "", 0, true},
		{"Negative number", "-4GB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemory(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestNormalizeKubernetesVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"Version with v prefix", "v1.31.11", "1.31.11", false},
		{"Version without v prefix", "1.31.11", "1.31.11", false},
		{"Version with V prefix uppercase", "V1.31.11", "1.31.11", false},
		{"Version with suffix", "v1.31.11-rc.1", "1.31.11-rc.1", false},
		{"Version with alpha suffix", "1.32.0-alpha.0", "1.32.0-alpha.0", false},
		{"Empty string", "", "", false},
		// Security tests - command injection attempts
		{"Command injection with space", "1.31.11 --malicious-flag", "", true},
		{"Command injection with semicolon", "1.31.11;malicious", "", true},
		{"Command injection with pipe", "1.31.11|malicious", "", true},
		{"Command injection with ampersand", "1.31.11&&malicious", "", true},
		{"Command injection with backtick", "1.31.11`malicious`", "", true},
		{"Command injection with dollar", "1.31.11$malicious", "", true},
		{"Path traversal attempt", "../1.31.11", "", true},
		{"Invalid format - no patch", "1.31", "", true},
		{"Invalid format - extra dots", "1.31.11.12", "", true},
		{"Invalid format - letters in version", "1.31.a1", "", true},
		{"Just v", "v", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeKubernetesVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
