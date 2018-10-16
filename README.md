Gobble
=======
An application deployment system written in Go, intended to leave as much configuration as possible to the developer.

## Installation
The following section describes how to install Gobble and its dependencies.

###### Dependencies
Gobble only has one true dependency: Golang. Docker is an option, and can be disabled with a command line flag. If you want to use the Docker deployment feature, make sure you have it installed by following the instructions [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce-1).

#### Setting up SSH
Docker uses SSH to pull and clone from git repositories. This can cause trouble if not set up correctly.

We'll start by creating a new SSH key and adding it to GitHub. If you've done this already, skip to step 3.
**Perform all of these steps as root.**

1. In a root terminal window, generate a new SSH key:
`$  ssh-keygen`. Use all of the default options.

2. Copy the contents of `~/.ssh/id_rsa.pub` into the text box for a new SSH key in the GitHub "SSH and GPG keys" page, which can be found under "Settings" for your account. Name it whatever you want, and click "Add SSH Key."

3. Back in a terminal, run the following:

```bash
$  eval `ssh-agent`
$  ssh-add
```

This will ensure you have an ssh-agent running, and that its identity is associated with the SSH keys you just created.


## Configuration

## Execution
