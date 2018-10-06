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
	Name       string
	killed     chan (bool)
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
		var container Container
		container.DeployContainer(*d)
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
