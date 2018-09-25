package utils

import "fmt"

const (
	GITPULL  = iota
	GITCLONE = iota
)

var GitActions = map[int]string{
	GITPULL:  "Git Pull",
	GITCLONE: "Git Clone",
}

type ERRGITWEBHOOK struct {
	GitAction int
	Message   string
}

func (e ERRGITWEBHOOK) Error() string {
	return fmt.Sprintf("%s: %s", GitActions[e.GitAction], e.Message)
}

// ===============================================================

type ErrNoConfig struct{}

var (
	ERRNOCONFIG = ErrNoConfig{}
)

func (e ErrNoConfig) Error() string {
	return "No config file found for this repository"
}
