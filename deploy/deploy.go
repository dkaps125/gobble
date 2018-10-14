package deploy

import (
	"context"
	"gobble/utils"
	"log"
)

var (
	platforms = map[string]string{
		"golang": "/go/src",
		"node":   "/deploy",
	}
)

type Deploy struct {
	Name string

	DeployType string            `json:"type"`
	Build      string            `json:"build"`
	Test       string            `json:"test"`
	Run        string            `json:"run"`
	Platform   string            `json:"platform"`
	Ports      map[string]string `json:"ports"`

	runCancel context.CancelFunc
	killed    chan (bool)
}

var deployments = make(map[string]*Deploy)

func (d *Deploy) Deploy(name string) error {
	d.Name = name

	log.Printf("Checking for prior deployment of '%s'\n", name)
	if dep, ok := deployments[name]; ok {
		log.Printf("Found previous deployment of %s. Stopping...\n", name)
		dep.runCancel()
		died := <-dep.killed

		if !died {
			return utils.ERRKILLPROC
		}

		delete(deployments, name)
		log.Printf("Stopped prior deployment of %s\n", name)
	}

	if d.DeployType == "local" {
		err := d.build()
		if err != nil {
			return err
		}

		err = d.test()
		if err != nil {
			return err
		}

		err = d.run()
		if err != nil {
			return err
		}
	} else if d.DeployType == "docker" {
		if utils.Config.NoDocker {
			return utils.ERRNOCONFIG
		}

		if _, ok := platforms[d.Platform]; !ok {
			return utils.ERRINVALIDPLATFORM
		}

		log.Printf("Deploying %s in a new docker container\n", d.Name)
		var container Container
		err := container.DeployContainer(*d)

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Deploy) build() error {
	if d.Build != "" {
		_, _, err := ExecuteCommand(d.Build, utils.Config.Timeout)
		return err
	}

	return nil
}

func (d *Deploy) test() error {
	if d.Test != "" {
		_, _, err := ExecuteCommand(d.Test, utils.Config.Timeout)
		return err
	}

	return nil
}

func (d *Deploy) run() error {
	if d.Run != "" {
		cancel, killed, err := ExecuteCommand(d.Run, 0)
		d.runCancel = cancel
		d.killed = killed

		if err == nil {
			log.Printf("Launched new version of %s\n", d.Name)
			deployments[d.Name] = d
		}

		return err
	}

	return nil
}

func Shutdown() {
	for _, v := range deployments {
		log.Printf("Shutting down %s\n", v.Name)
		v.runCancel()
		<-v.killed
	}
}
