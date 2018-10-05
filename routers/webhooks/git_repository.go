package webhooks

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"gopkg.in/src-d/go-git.v4"

	"gobble/deploy"
	"gobble/utils"
)

type Repo struct {
	Name      string `json:"name"`
	Clone_url string `json:"clone_url"`
	Git_url   string `json:"git_url"`
	Ssh_url   string `json:"ssh_url"`
	directory string
	config    ProjConfig
}

type ProjConfig struct {
	Name       string        `json:"name"`
	Deployment deploy.Deploy `json:"deploy"`
}

func (r *Repo) SetDirectory(directory string) {
	r.directory = directory
}

func (r *Repo) UpdateOrClone() error {
	if r.repoDirectoryExists() {
		repo, err := git.PlainOpen(r.directory)

		if err != nil {
			return utils.ERRGITWEBHOOK{
				GitAction: utils.GITPULL,
				Message:   "Repository directory not found",
			}
		}

		worktree, err := repo.Worktree()

		if err != nil {
			return utils.ERRGITWEBHOOK{
				GitAction: utils.GITPULL,
				Message:   "Unable to import worktree of existing repository",
			}
		}

		//TODO: error checking
		worktree.Pull(&git.PullOptions{
			SingleBranch: true,
		})
	} else {
		log.Printf("Cloning from %s\n", r.Clone_url)
		//TODO: use config to determine which url to clone from
		_, err := git.PlainClone(r.directory, false, &git.CloneOptions{
			URL: string(r.Ssh_url),
		})

		if err != nil {
			fmt.Println(err)
			return utils.ERRGITWEBHOOK{
				GitAction: utils.GITCLONE,
				Message:   "Repository could not be cloned",
			}
		}
	}

	return nil
}

//TODO: hash config file to see if it needs to be reprocessed
func (r *Repo) ImportConfig() error {
	configPath := path.Join(r.directory, "gobble.json")

	configFile, err := os.Open(configPath)
	defer configFile.Close()

	if err != nil {
		return utils.ERRNOCONFIG
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&r.config)

	if r.config.Name != r.Name {
		return utils.ERRNOCONFIG
	}

	return nil
}

func (r *Repo) Deploy() error {
	err := os.Chdir(path.Join(utils.Config.WorkingDir, r.directory))
	defer os.Chdir(utils.Config.WorkingDir)

	if err != nil {
		return err
	}

	//TODO: do i need to error check here?
	r.config.Deployment.Deploy(r.Name)

	return nil
}

func (r *Repo) repoDirectoryExists() bool {
	return utils.DirectoryExists(r.directory)
}
