package deploy

import (
	"context"
	"log"
	"os"
	"path"
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
	dir    string
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

	dockerClient = cli
	dockerContext = ctx

	log.Println("Docker interface initalized")
}

func (container *Container) DeployContainer(dep Deploy) error {
	container.state = NEW
	container.deploy = dep

	log.Println("Pulling docker image")
	_, err := dockerClient.ImagePull(dockerContext, "docker.io/library/"+dep.Platform, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	//io.Copy(os.Stdout, reader)

	err = container.initContainer()
	log.Println("Container initialized")
	if err != nil {
		return err
	}

	err = container.deployArchive()
	log.Println("Deployed project archive to container")
	if err != nil {
		return err
	}

	err = container.startContainer()
	log.Println("Started container")
	if err != nil {
		return err
	}

	err = container.deployContainer()
	log.Println("Launched project on container")
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) initContainer() error {
	if c.state != NEW {
		return utils.ERRINVALIDSTATE
	}

	exposed, bound := getPorts(c.deploy)

	dir := path.Join(platforms[c.deploy.Platform], c.deploy.Name)

	resp, err := dockerClient.ContainerCreate(dockerContext, &container.Config{
		Image:        c.deploy.Platform,
		WorkingDir:   dir,
		Entrypoint:   []string{"/bin/bash"},
		Tty:          true,
		ExposedPorts: exposed,
	}, &container.HostConfig{
		PortBindings: bound,
	}, nil, "")

	if err != nil {
		return err
	}

	c.id = resp.ID
	c.state = CREATED
	c.dir = dir
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

	if err := dockerClient.CopyToContainer(dockerContext, c.id, c.dir, archive, types.CopyToContainerOptions{}); err != nil {
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
		Cmd:        strings.Split(command, " "),
		WorkingDir: c.dir,
		Tty:        true,
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

func getPorts(dep Deploy) (nat.PortSet, nat.PortMap) {

	exposed := make(nat.PortSet)
	bound := make(nat.PortMap)

	for k, v := range dep.Ports {
		exposed[nat.Port(k)] = struct{}{}

		bound[nat.Port(k)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: v,
			},
		}
	}

	return exposed, bound
}
