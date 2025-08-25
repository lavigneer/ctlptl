package rancher

import (
	"strings"

	"github.com/tilt-dev/ctlptl/pkg/docker"
)

const (
	AdminSocket    = "/var/run/docker.sock"
	NonAdminSocket = "/.rd/docker.sock"
)

// IsLocalRancherEngineHost checks whether the DOCKER_HOST looks like a local Rancher Desktop Engine.
func IsLocalRancherEngineHost(dockerHost string) bool {
	if strings.HasPrefix(dockerHost, "unix:") {
		return strings.Contains(dockerHost, AdminSocket) ||
			// Rancher Desktop for non-admin - socket is in ~/.rd/docker.sock
			strings.HasSuffix(dockerHost, NonAdminSocket)
	}

	// Docker daemons on other local protocols are treated as rancher desktop.
	return docker.IsLocalHost(dockerHost)
}

// IsLocalRancherDesktop checks whether the DOCKER_HOST looks like a local Rancher Desktop.
func IsLocalRancherDesktop(dockerHost string, os string) bool {
	switch os {
	case "windows":
		return docker.IsLocalHost(dockerHost)
	case "darwin":
		return IsLocalRancherEngineHost(dockerHost)
	case "linux":
		return strings.HasPrefix(dockerHost, "unix:") &&
			strings.HasSuffix(dockerHost, NonAdminSocket)
	}
	return false
}
