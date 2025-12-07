package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	klog "k8s.io/klog/v2"

	"github.com/tilt-dev/ctlptl/pkg/api"
)

// RancherManager interface defines the operations for managing Rancher Desktop
type RancherManager interface {
	// Status gets the current status of Rancher Desktop
	Status(ctx context.Context) (map[string]interface{}, error)
	
	// Start starts Rancher Desktop with the desired configuration
	Start(ctx context.Context, cluster *api.Cluster) error
	
	// ResetCluster resets the Kubernetes cluster
	ResetCluster(ctx context.Context) error
	
	// Shutdown stops Rancher Desktop
	Shutdown(ctx context.Context) error
	
	// IsKubernetesRunning checks if Kubernetes is running in Rancher Desktop
	IsKubernetesRunning(ctx context.Context) (bool, error)
}

// RancherDesktopManager manages Rancher Desktop using rdctl commands.
// This implementation focuses on leveraging rdctl's efficient startup flags 
// rather than multiple separate setting adjustments.
type RancherDesktopManager struct{}

// NewRancherDesktopManager creates a new manager for Rancher Desktop.
func NewRancherDesktopManager() (RancherManager, error) {
	// Verify rdctl is installed
	_, err := exec.LookPath("rdctl")
	if err != nil {
		return nil, fmt.Errorf("Rancher Desktop not installed: rdctl command not found")
	}
	return &RancherDesktopManager{}, nil
}

// Status gets the current status of Rancher Desktop
func (r *RancherDesktopManager) Status(ctx context.Context) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "rdctl", "api", "status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Rancher Desktop status: %v", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse Rancher Desktop status: %v", err)
	}
	
	return status, nil
}

// Start starts Rancher Desktop with the desired configuration
func (r *RancherDesktopManager) Start(ctx context.Context, cluster *api.Cluster) error {
	// Build command with options
	args := []string{"start"}

	// Always enable Kubernetes for ctlptl
	args = append(args, "--kubernetes-enabled=true")

	// Apply CPU settings
	if cluster.MinCPUs > 0 {
		args = append(args, fmt.Sprintf("--virtual-machine.number-cpus=%d", cluster.MinCPUs))
	}

	// Apply memory settings
	if cluster.Memory != "" {
		memGB, err := parseMemory(cluster.Memory)
		if err != nil {
			return err
		}
		args = append(args, fmt.Sprintf("--virtual-machine.memory-in-gb=%d", memGB))
	}

	// Apply Kubernetes version if specified
	// Rancher Desktop expects versions without 'v' prefix
	if cluster.KubernetesVersion != "" {
		version := normalizeKubernetesVersion(cluster.KubernetesVersion)
		args = append(args, fmt.Sprintf("--kubernetes-version=%s", version))
	}

	// Add background option to avoid blocking
	args = append(args, "--application.start-in-background")
	
	klog.V(2).Infof("Starting Rancher Desktop with args: %v", args)
	cmd := exec.CommandContext(ctx, "rdctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start Rancher Desktop: %v, output: %s", err, string(output))
	}
	
	// Wait for Kubernetes to be ready
	return r.waitForKubernetes(ctx)
}

// ResetCluster resets the Kubernetes cluster
func (r *RancherDesktopManager) ResetCluster(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "rdctl", "factory-reset", "--kubernetes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reset Rancher Desktop Kubernetes: %v, output: %s", err, string(output))
	}
	
	// Wait for Kubernetes to be ready again
	return r.waitForKubernetes(ctx)
}

// Shutdown stops Rancher Desktop
func (r *RancherDesktopManager) Shutdown(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "rdctl", "shutdown")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to shutdown Rancher Desktop: %v, output: %s", err, string(output))
	}
	return nil
}

// Note: SetSetting and SetKubernetesVersion methods removed as they're not used

// IsKubernetesRunning checks if Kubernetes is running in Rancher Desktop
func (r *RancherDesktopManager) IsKubernetesRunning(ctx context.Context) (bool, error) {
	status, err := r.Status(ctx)
	if err != nil {
		return false, err
	}
	
	// Navigate nested maps to find kubernetes status
	vmMap, ok := status["vm"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	
	kubeMap, ok := vmMap["kubernetes"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	
	state, ok := kubeMap["state"].(string)
	return ok && state == "running", nil
}

// parseMemory converts memory strings like "4GB", "8GB", "2048MB" to GB as an integer
func parseMemory(memory string) (int, error) {
	memory = strings.TrimSpace(strings.ToUpper(memory))

	if strings.HasSuffix(memory, "GB") {
		gb, err := strconv.Atoi(strings.TrimSuffix(memory, "GB"))
		if err != nil {
			return 0, fmt.Errorf("invalid memory format: %s", memory)
		}
		if gb <= 0 {
			return 0, fmt.Errorf("memory must be positive: %s", memory)
		}
		return gb, nil
	}

	if strings.HasSuffix(memory, "MB") {
		mb, err := strconv.Atoi(strings.TrimSuffix(memory, "MB"))
		if err != nil {
			return 0, fmt.Errorf("invalid memory format: %s", memory)
		}
		if mb <= 0 {
			return 0, fmt.Errorf("memory must be positive: %s", memory)
		}
		// Convert MB to GB, rounding up
		gb := (mb + 1023) / 1024
		return gb, nil
	}

	// Try parsing as a plain number (assume GB)
	gb, err := strconv.Atoi(memory)
	if err != nil {
		return 0, fmt.Errorf("invalid memory format: %s (expected formats: 4GB, 8GB, 2048MB)", memory)
	}
	if gb <= 0 {
		return 0, fmt.Errorf("memory must be positive: %s", memory)
	}
	return gb, nil
}

// normalizeKubernetesVersion strips the 'v' or 'V' prefix from Kubernetes version if present
// Rancher Desktop expects versions without the 'v' prefix
func normalizeKubernetesVersion(version string) string {
	if len(version) > 0 && (version[0] == 'v' || version[0] == 'V') {
		return version[1:]
	}
	return version
}

// waitForKubernetes waits for Kubernetes to be ready
func (r *RancherDesktopManager) waitForKubernetes(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(1 * time.Minute)
	
	for {
		select {
		case <-ticker.C:
			if running, err := r.IsKubernetesRunning(ctx); err == nil && running {
				return nil
			}
			klog.V(2).Info("Waiting for Kubernetes to be ready in Rancher Desktop...")
			
		case <-timeout:
			return errors.New("timed out waiting for Kubernetes to be ready in Rancher Desktop")
			
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
