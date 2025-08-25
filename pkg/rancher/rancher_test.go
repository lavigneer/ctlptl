package rancher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type rancherDesktopTestCase struct {
	host     string
	os       string
	expected bool
}

func TestIsLocalRancherDesktop(t *testing.T) {
	cases := []rancherDesktopTestCase{
		{"", "linux", false},
		{"tcp://localhost:2375", "linux", false},
		{"tcp://127.0.0.1:2375", "linux", false},
		{"npipe:////./pipe/docker_engine", "windows", true},
		{"unix:///var/run/docker.sock", "darwin", true},
		{"unix:///var/run/docker.sock", "linux", false},
		{"tcp://cluster:2375", "linux", false},
		{"http://cluster:2375", "linux", false},
		{"unix:///Users/USER/.rd/docker.sock", "linux", true},
		{"unix:///Users/USER/.rd/docker.sock", "darwin", true},
		{"unix:///Users/USER/.docker/desktop/docker.sock", "linux", false},
	}
	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%s-%d", t.Name(), i), func(t *testing.T) {
			assert.Equal(t, c.expected, IsLocalRancherDesktop(c.host, c.os))
		})
	}
}
