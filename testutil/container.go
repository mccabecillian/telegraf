//go:build !freebsd
// +build !freebsd

package testutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Container struct {
	BindMounts   map[string]string
	Entrypoint   []string
	Env          map[string]string
	ExposedPorts []string
	Image        string
	Name         string
	Networks     []string
	WaitingFor   wait.Strategy

	Address string
	Ports   map[string]string

	container testcontainers.Container
	ctx       context.Context
}

func (c *Container) Start() error {
	c.ctx = context.Background()

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			BindMounts:   c.BindMounts,
			Entrypoint:   c.Entrypoint,
			Env:          c.Env,
			ExposedPorts: c.ExposedPorts,
			Image:        c.Image,
			Name:         c.Name,
			Networks:     c.Networks,
			WaitingFor:   c.WaitingFor,
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(c.ctx, req)
	if err != nil {
		return fmt.Errorf("container failed to start: %s", err)
	}
	c.container = container

	c.Address = "localhost"

	err = c.LookupMappedPorts()
	if err != nil {
		_ = c.Terminate()
		return fmt.Errorf("port lookup failed: %s", err)
	}

	return nil
}

// create a lookup table of exposed ports to mapped ports
func (c *Container) LookupMappedPorts() error {
	if len(c.ExposedPorts) == 0 {
		return nil
	}

	if len(c.Ports) == 0 {
		c.Ports = make(map[string]string)
	}

	for _, port := range c.ExposedPorts {
		// strip off leading host port: 80:8080 -> 8080
		if strings.Contains(port, ":") {
			port = strings.Split(port, ":")[1]
		}

		// strip off the transport: 80/tcp -> 80
		if strings.Contains(port, "/") {
			port = strings.Split(port, "/")[0]
		}

		p, err := c.container.MappedPort(c.ctx, nat.Port(port))
		if err != nil {
			return fmt.Errorf("failed to find '%s' - %s", port, err)
		}
		fmt.Printf("mapped container port '%s' to host port '%s'\n", port, p.Port())
		c.Ports[port] = p.Port()
	}

	return nil
}

func (c *Container) Terminate() error {
	err := c.container.Terminate(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to terminate the container: %s", err)
	}

	return nil
}
