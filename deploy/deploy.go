package deploy

import (
	"gobble/utils"
	"log"
	"os"
	"path"
)

var (
	platforms = map[string]string{
		"golang": "/go/src",
		"node":   "/deploy",
	}
)

type DeploymentType interface {
	Deploy(dep Deploy) error
	Halt() error
}

type Deploy struct {
	Name string

	DeployType string            `json:"type"`
	Build      string            `json:"build"`
	Test       string            `json:"test"`
	Run        string            `json:"run"`
	Platform   string            `json:"platform"`
	Ports      map[string]string `json:"ports"`

	pidFile string
}

var deployments = make(map[string]DeploymentType)

func (d *Deploy) Deploy(name string) error {
	d.Name = name

	log.Printf("Checking for prior deployment of '%s'\n", name)
	err := d.removePrev()
	if err != nil {
		return err
	}

	if d.DeployType == "local" {
		log.Printf("Deploying %s locally\n", d.Name)

		var local Local
		err := local.Deploy(*d)

		if err != nil {
			return err
		}

		deployments[d.Name] = &local
	} else if d.DeployType == "docker" {
		if utils.Config.NoDocker {
			return utils.ERRNOCONFIG
		}

		if _, ok := platforms[d.Platform]; !ok {
			return utils.ERRINVALIDPLATFORM
		}

		log.Printf("Deploying %s in a new docker container\n", d.Name)
		var container Container
		err := container.Deploy(*d)

		if err != nil {
			return err
		}

		deployments[d.Name] = &container
	}

	return d.createPID()
}

func (d *Deploy) removePrev() error {
	name := d.Name

	if dep, ok := deployments[name]; ok {
		log.Printf("Found previous deployment of %s. Stopping...\n", name)

		err := dep.Halt()
		if err != nil {
			return err
		}

		delete(deployments, name)
		log.Printf("Stopped prior deployment of %s\n", name)

		return d.removePID()
	}

	return nil
}

func (d *Deploy) removePID() error {
	err := os.Remove(d.pidFile)

	if err == nil {
		d.pidFile = ""
	}

	return err
}

func (d *Deploy) createPID() error {
	pidFile := path.Join(utils.Config.GetPidDir(), d.Name+".pid")
	_, err := os.Create(pidFile)

	if err == nil {
		d.pidFile = pidFile
		log.Printf("Created new pidFile %s\n", d.pidFile)
	}

	return err
}

func Shutdown() {
	for k, v := range deployments {
		log.Printf("Shutting down %s\n", k)

		err := v.Halt()
		if err != nil {
			log.Printf("Deployment %s may not have shut down correctly\n", k)
		}

	}
}
