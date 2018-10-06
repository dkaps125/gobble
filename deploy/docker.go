package deploy

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"gobble/utils"
)

const (
	NEW      = iota
	CREATED  = iota
	SHIPPED  = iota
	STARTED  = iota
	DEPLOYED = iota
	STOPPED  = iota
)

type Container struct {
	state  int
	deploy Deploy
	id     string
}

var dockerClient *client.Client
var dockerContext context.Context

func InitDocker() {
	log.Println("Initializing Docker interface...")
	ctx := context.Background()
	cli, err := client.NewEnvClient()

	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/ubuntu", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	dockerClient = cli
	dockerContext = ctx

	log.Println("Docker interface initalized, ubuntu image imported")
}

func (container *Container) DeployContainer(dep Deploy) error {
	container.state = NEW
	container.deploy = dep

	err := container.initContainer()
	if err != nil {
		return err
	}

	err = container.deployArchive()
	if err != nil {
		return err
	}

	err = container.startContainer()
	if err != nil {
		return err
	}

	err = container.deployContainer()
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) initContainer() error {
	if c.state != NEW {
		return utils.ERRINVALIDSTATE
	}

	resp, err := dockerClient.ContainerCreate(dockerContext, &container.Config{
		Image:        "ubuntu",
		WorkingDir:   "/deploy",
		Entrypoint:   []string{"/bin/bash"},
		Tty:          true,
		ExposedPorts: nat.PortSet{}, //TODO: obtain this
	}, &container.HostConfig{
		PortBindings: nat.PortMap{}, //TODO: obtain this
	}, nil, "")

	if err != nil {
		return err
	}

	c.id = resp.ID
	c.state = CREATED
	return nil
}

func (c *Container) deployArchive() error {
	if c.state != CREATED {
		return utils.ERRINVALIDSTATE
	}

	archiveName, err := utils.Tar(c.deploy.Name)
	if err != nil {
		return err
	}

	archive, err := os.Open(archiveName)
	if err != nil {
		return err
	}

	if err := dockerClient.CopyToContainer(dockerContext, c.id, "/deploy", archive, types.CopyToContainerOptions{}); err != nil {
		return err
	}

	c.state = SHIPPED
	return nil
}

func (c *Container) startContainer() error {
	if c.state != SHIPPED {
		return utils.ERRINVALIDSTATE
	}

	if err := dockerClient.ContainerStart(dockerContext, c.id, types.ContainerStartOptions{}); err != nil {
		return err
	}

	c.state = STARTED
	return nil
}

func (c *Container) deployContainer() error {
	done, err := c.runBlockingCommand(c.deploy.Build)
	if err != nil {
		return err
	}

	<-done

	done, err = c.runBlockingCommand(c.deploy.Test)
	if err != nil {
		return err
	}

	<-done

	err = c.runNonblockingCommand(c.deploy.Run)
	if err != nil {
		return err
	}

	c.state = DEPLOYED
	return nil
}

func (c *Container) runBlockingCommand(command string) (chan (int), error) {
	return c.runCommand(command, true)
}

func (c *Container) runNonblockingCommand(command string) error {
	_, err := c.runCommand(command, false)

	return err
}

func (c *Container) runCommand(command string, block bool) (chan (int), error) {
	if c.state != STARTED {
		return nil, utils.ERRINVALIDSTATE
	}

	execId, err := dockerClient.ContainerExecCreate(dockerContext, c.id, types.ExecConfig{
		Cmd: strings.Split(command, " "),
		Tty: true,
	})

	if err != nil {
		return nil, err
	}

	if err := dockerClient.ContainerExecStart(dockerContext, execId.ID, types.ExecStartCheck{}); err != nil {
		return nil, err
	}

	done := make(chan (int), 1)

	if block {

		go func() {
			for {
				resp, err := dockerClient.ContainerExecInspect(dockerContext, execId.ID)
				if err != nil {
					log.Printf("Container command check failed: %v\n", err)
					return
				}

				if !resp.Running {
					done <- 1
					break
				}

				time.Sleep(5 * time.Second)
			}
		}()
	}

	return done, nil
}
