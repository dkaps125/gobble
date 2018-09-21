package utils

import (
	"os"

	"golang.org/x/sys/unix"
)

type Configuration struct {
	Port       int
	projectDir string
}

var Config Configuration

func (c *Configuration) SetProjectDir(directory string) {
	var err error

	if DirectoryExists(directory) {
		err = unix.Access(directory, unix.W_OK)
	} else {
		err = os.Mkdir(directory, os.ModeDir|0776)
	}

	if err != nil {
		panic("Could not set/create project directory!\n\n" + err.Error())
	}

	c.projectDir = directory
}

func (c *Configuration) GetProjectDir() string {
	return c.projectDir
}
