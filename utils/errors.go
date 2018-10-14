package utils

import "fmt"

const (
	GITHOOK  = iota
	GITPULL  = iota
	GITCLONE = iota
)

var GitActions = map[int]string{
	GITHOOK:  "Git Hook",
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

type errFile struct {
	Message string
}

var (
	ERRNOCONFIG = errFile{
		Message: "No config file found for this repository",
	}

	ERRFILENOTFOUND = errFile{
		Message: "File not found",
	}

	ERRNOOPEN = errFile{
		Message: "Unable to open file",
	}
)

func (e errFile) Error() string {
	return e.Message
}

// ===============================================================

const (
	errKillProc = iota
)

var deployMessages = map[int]string{
	errKillProc: "Error killing process",
}

type errDeploy struct {
	DeployFailure int
}

func (e errDeploy) Error() string {
	return deployMessages[e.DeployFailure]
}

var (
	ERRKILLPROC = errDeploy{
		DeployFailure: errKillProc,
	}
)

// ===============================================================

const (
	errInvalidState    = iota
	errInvalidPlatform = iota
)

var dockerMessages = map[int]string{
	errInvalidState:    "Container was in an invalid state for called operation",
	errInvalidPlatform: "Invalid container platform was specified",
}

type dockerError struct {
	ErrorType int
}

func (e dockerError) Error() string {
	return dockerMessages[e.ErrorType]
}

var (
	ERRINVALIDSTATE = dockerError{
		ErrorType: errInvalidState,
	}

	ERRINVALIDPLATFORM = dockerError{
		ErrorType: errInvalidPlatform,
	}
)
