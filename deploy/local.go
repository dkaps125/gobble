package deploy

import (
	"context"
	"log"

	"gobble/utils"
)

type Local struct {
	deploy Deploy

	runCancel context.CancelFunc
	killed    chan (bool)
}

func (l *Local) Deploy(dep Deploy) error {
	l.deploy = dep

	err := l.build()
	if err != nil {
		return err
	}

	err = l.test()
	if err != nil {
		return err
	}

	err = l.run()

	return err
}

func (l *Local) Halt() error {
	l.runCancel()
	died := <-l.killed

	if !died {
		return utils.ERRKILLPROC
	}

	return nil
}

func (l *Local) build() error {
	if l.deploy.Build != "" {
		_, _, err := ExecuteCommand(l.deploy.Build, utils.Config.Timeout)
		return err
	}

	return nil
}

func (l *Local) test() error {
	if l.deploy.Test != "" {
		_, _, err := ExecuteCommand(l.deploy.Test, utils.Config.Timeout)
		return err
	}

	return nil
}

func (l *Local) run() error {
	if l.deploy.Run != "" {
		cancel, killed, err := ExecuteCommand(l.deploy.Run, 0)
		l.runCancel = cancel
		l.killed = killed

		if err == nil {
			log.Printf("Launched new version of %s\n", l.deploy.Name)
		}

		return err
	}

	return nil
}
