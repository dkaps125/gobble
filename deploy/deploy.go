package deploy

import (
	"context"
	"gobble/utils"
	"log"
)

type Deploy struct {
	DeployType string `json:"type"`
	Build      string `json:"build"`
	Test       string `json:"test"`
	Run        string `json:"run"`
	runCancel  context.CancelFunc
	name       string
}

var deployments = make(map[string]*Deploy)

func (d *Deploy) Deploy(name string) error {
	d.name = name

	log.Printf("Checking for prior deployment of %s\n", name)
	if dep, ok := deployments[name]; ok {
		log.Printf("Found previous deployment of %s. Stopping...\n", name)
		dep.runCancel()
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
	} else if d.DeployType == "container" {
		//TODO: deploy with docker container
	}

	return nil
}

func (d *Deploy) build() error {
	if d.Build != "" {
		_, err := utils.ExecuteCommand(d.Build, utils.Config.Timeout)
		return err
	}

	return nil
}

func (d *Deploy) test() error {
	if d.Test != "" {
		_, err := utils.ExecuteCommand(d.Test, utils.Config.Timeout)
		return err
	}

	return nil
}

func (d *Deploy) run() error {
	if d.Run != "" {
		cancel, err := utils.ExecuteCommand(d.Run, 0)
		d.runCancel = cancel

		if err == nil {
			deployments[d.name] = d
		}

		return err
	}

	return nil
}

func Shutdown() {
	for _, v := range deployments {
		v.runCancel()
	}
}
