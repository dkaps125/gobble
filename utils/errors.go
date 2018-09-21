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
