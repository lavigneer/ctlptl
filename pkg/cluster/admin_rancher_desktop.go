package cluster

import (
	"context"
	"fmt"

	"github.com/tilt-dev/localregistry-go"

	"github.com/tilt-dev/ctlptl/pkg/api"
	"github.com/tilt-dev/ctlptl/pkg/rancher"
)

type rancherDesktopAdmin struct {
	os     string
	host   string
	rdm    RancherManager
}

func newRancherDesktopAdmin(host string, os string, rdm RancherManager) *rancherDesktopAdmin {
	return &rancherDesktopAdmin{os: os, host: host, rdm: rdm}
}

func (a *rancherDesktopAdmin) EnsureInstalled(ctx context.Context) error {
	return nil
}

func (a *rancherDesktopAdmin) Create(ctx context.Context, desired *api.Cluster, registry *api.Registry) error {
	// TODO: Implement registry support for Rancher Desktop
	if registry != nil {
		return fmt.Errorf("ctlptl currently does not support connecting a registry to rancher-desktop")
	}
	if len(desired.RegistryAuths) > 0 {
		return fmt.Errorf("ctlptl currently does not support connecting pull-through registries to rancher-desktop")
	}

	isLocalRancherDesktop := rancher.IsLocalRancherDesktop(a.host, a.os)
	if !isLocalRancherDesktop {
		return fmt.Errorf("rancher-desktop clusters are only available on a local Rancher Desktop. Current DOCKER_HOST: %s",
			a.host)
	}

	err := a.rdm.Start(ctx, desired)
	if err != nil {
		return fmt.Errorf("failed to start Rancher Desktop: %v", err)
	}

	return nil
}

func (a *rancherDesktopAdmin) LocalRegistryHosting(ctx context.Context, desired *api.Cluster, registry *api.Registry) (*localregistry.LocalRegistryHostingV1, error) {
	return nil, nil
}

func (a *rancherDesktopAdmin) Delete(ctx context.Context, config *api.Cluster) error {
	isLocalRancherDesktop := rancher.IsLocalRancherDesktop(a.host, a.os)
	if !isLocalRancherDesktop {
		return fmt.Errorf("rancher-desktop cannot be deleted from DOCKER_HOST: %s", a.host)
	}

	return a.rdm.ResetCluster(ctx)
}
