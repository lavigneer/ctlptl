# Rancher Desktop Support

ctlptl supports creating and managing Kubernetes clusters using Rancher Desktop, a desktop application that provides Kubernetes and container management capabilities.

## Prerequisites

1. **Rancher Desktop**: Install Rancher Desktop from [https://rancherdesktop.io/](https://rancherdesktop.io/)
2. **rdctl**: The Rancher Desktop command-line tool should be available in your PATH

## Basic Usage

Create a simple Rancher Desktop cluster:

```bash
ctlptl create cluster rancher-desktop
```

Or use a configuration file:

```yaml
apiVersion: ctlptl.dev/v1alpha1
kind: Cluster
name: rancher-desktop
product: rancher-desktop
minCPUs: 4
memory: 8GB
kubernetesVersion: v1.31.11
```

Apply the configuration:

```bash
ctlptl apply -f cluster.yaml
```

## Configuration Options

### Basic Configuration

- `name`: The name of the cluster (defaults to "rancher-desktop")
- `product`: Must be set to "rancher-desktop"
- `minCPUs`: Minimum number of CPUs to allocate to the VM (default: 2)
- `memory`: Amount of RAM to allocate to the VM (e.g., "4GB", "8GB", "2048MB")
- `kubernetesVersion`: Kubernetes version to use (with or without 'v' prefix)

### Advanced Configuration

```yaml
apiVersion: ctlptl.dev/v1alpha1
kind: Cluster
name: rancher-desktop-advanced
product: rancher-desktop
minCPUs: 6
memory: 8GB
kubernetesVersion: v1.31.11
```

### Supported Options

- **minCPUs**: Configures the number of CPUs allocated to the Rancher Desktop VM
- **memory**: Configures the amount of RAM allocated to the Rancher Desktop VM
  - Supported formats: `4GB`, `8GB`, `2048MB` (plain numbers assumed to be GB)
  - Example: `memory: 8GB`
- **kubernetesVersion**: Specifies the Kubernetes version to use
  - Accepts versions with or without 'v' prefix (e.g., "v1.31.11" or "1.31.11")
  - The 'v' prefix will be automatically stripped as Rancher Desktop expects versions without it

### Unsupported Options

Currently, the following features are not supported for Rancher Desktop clusters:

- **Registry**: Connecting local registries to Rancher Desktop clusters
- **RegistryAuths**: Pull-through registry authentication
- **Custom cluster configurations**: Rancher Desktop uses a simplified configuration model

## Management Commands

### List Clusters

```bash
ctlptl get clusters
```

### Get Cluster Status

```bash
ctlptl get cluster rancher-desktop
```

### Delete Cluster

```bash
ctlptl delete cluster rancher-desktop
```

## How It Works

ctlptl integrates with Rancher Desktop using the `rdctl` command-line tool to:

1. **Start Rancher Desktop**: Automatically starts Rancher Desktop with the specified configuration
2. **Configure Kubernetes**: Enables Kubernetes and sets the desired version
3. **Manage Resources**: Configures CPU and memory allocation for the VM
4. **Monitor Status**: Tracks the cluster's health and status

## Troubleshooting

### Rancher Desktop Not Found

If you see an error about `rdctl` not being found:

1. Ensure Rancher Desktop is installed
2. Verify that `rdctl` is in your PATH
3. Try restarting your terminal

### Cluster Won't Start

If the cluster fails to start:

1. Check that Rancher Desktop is not already running with different settings
2. Verify that the specified Kubernetes version is available in Rancher Desktop
3. Check the Rancher Desktop logs for more details

### Connection Issues

If you can't connect to the cluster:

1. Ensure Rancher Desktop is running
2. Check that Kubernetes is enabled in Rancher Desktop
3. Verify your kubeconfig is pointing to the correct cluster

## Examples

See the `examples/` directory for complete configuration examples:

- `examples/rancher-desktop.yaml` - Basic configuration
- `examples/rancher-desktop-advanced.yaml` - Advanced configuration with custom settings

## Limitations

- Rancher Desktop clusters run locally and don't support remote Docker hosts
- Registry integration is not yet implemented
- Some advanced Kubernetes configurations may not be supported
- The cluster name is always "rancher-desktop" regardless of the configuration
