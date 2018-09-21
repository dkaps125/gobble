package routers

import (
  "fmt"
  "path"

  "gopkg.in/src-d/go-git.v4"

  "gobble/utils"
)

type Repo struct {
  Name string `json:"name"`
  Clone_url string `json:"clone_url"`
  Git_url string `json:"git_url"`
  Ssh_url string `json:"ssh_url"`
  Directory string
}

type WebhookData struct {
  Repository Repo `json:"repository"`
}

func (w *WebhookData) Configure() {
  w.Repository.Directory = path.Join(utils.Config.GetProjectDir(), w.Repository.Name)
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

func (r *Repo) repoDirectoryExists() bool {
  return utils.DirectoryExists(r.Directory)
}
