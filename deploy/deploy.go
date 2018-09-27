package deploy

import (
	"context"
	"fmt"
	"gobble/utils"
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

	fmt.Printf("Attempting deploy of %s\n%v\n", name, d)

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
