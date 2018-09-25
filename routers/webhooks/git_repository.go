package webhooks

import (
  "fmt"
  "path"
  "os"
  "encoding/json"

  "gopkg.in/src-d/go-git.v4"

  "gobble/utils"
)

type Repo struct {
  Name string `json:"name"`
  Clone_url string `json:"clone_url"`
  Git_url string `json:"git_url"`
  Ssh_url string `json:"ssh_url"`
  Directory string
  Config ProjConfig
}

type ProjConfig struct {
  Name string `json:"name"`
  Build string `json:"build"`
  BuildTimeout int `json:"buildTimeout"`
  Test string `json:"test"`
  TestTimeout int `json:"testTimeout"`
  Run string `json:"run"`
  RunTimeout int `json:"runTimeout"`
}

func (r *Repo) UpdateOrClone() error {
  if r.repoDirectoryExists() {
    repo, err := git.PlainOpen(r.Directory)

    if err != nil {
      return utils.ERRGITWEBHOOK{
        GitAction: utils.GITPULL,
        Message: "Repository directory not found",
      }
    }

    worktree, err := repo.Worktree()

    if err != nil {
      return utils.ERRGITWEBHOOK{
        GitAction: utils.GITPULL,
        Message: "Unable to import worktree of existing repository",
      }
    }

    worktree.Pull(&git.PullOptions{
      SingleBranch: true,
    })
  } else {
    fmt.Printf("Cloning into %s\n", r.Directory)
    //TODO: use config to determine which url to clone from
    _, err := git.PlainClone(r.Directory, false, &git.CloneOptions{
      URL: string(r.Clone_url),
    })

    if err != nil {
      return utils.ERRGITWEBHOOK{
        GitAction: utils.GITCLONE,
        Message: "Repository could not be cloned",
      }
    }
  }

  return nil
}

//TODO: hash config file to see if it needs to be reprocessed
func (r *Repo) ImportConfig() error {
  configPath := path.Join(r.Directory, "gobble.json")

  configFile, err := os.Open(configPath)
  defer configFile.Close()

  if err != nil {
    return utils.ERRNOCONFIG
  }

  jsonParser := json.NewDecoder(configFile)
  jsonParser.Decode(&r.Config)

  if r.Config.Name != r.Name {
    return utils.ERRNOCONFIG
  }

  return nil
}

func (r *Repo) Build() {
  r.executeConfigCommand(r.Config.Build, r.Config.BuildTimeout)
}

func (r *Repo) Test() {
  r.executeConfigCommand(r.Config.Test, r.Config.TestTimeout)
}

func (r *Repo) Run() {
  r.executeConfigCommand(r.Config.Run, r.Config.RunTimeout)
}

func (r *Repo) repoDirectoryExists() bool {
  return utils.DirectoryExists(r.Directory)
}

func (r *Repo) executeConfigCommand(command string, timeout int) error {
  if command != "" {

  }

  return nil
}
