package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func DirectoryExists(directory string) bool {
	if _, err := os.Stat(directory); err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}

func ExecuteCommand(command string, timeout int) (context.CancelFunc, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	toks := strings.Split(command, " ")
	name := toks[0]
	args := toks[1:]

	if timeout == 0 {
		ctx, cancel = context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, name, args...)
		CopyStderr(cmd)
		CopyStdout(cmd)

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		return cancel, nil
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
		defer cancel()

		cmd := exec.CommandContext(ctx, name, args...)
		CopyStderr(cmd)
		CopyStdout(cmd)

		if err := cmd.Start(); err != nil {
			fmt.Println("Failed starting command")
		}

		err := cmd.Wait()

		if ctx.Err() == context.DeadlineExceeded {
			return nil, ctx.Err()
		} else if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func CopyStdout(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed capturing command stdout")
	}

	go io.Copy(os.Stdout, stdout)
}

func CopyStderr(cmd *exec.Cmd) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Failed capturing command stderr")
	}

	go io.Copy(os.Stderr, stderr)
}

func HTTPErrorCheck(err error, w http.ResponseWriter, errorCode int) {
	if err != nil {
		http.Error(w, err.Error(), errorCode)
	}
}
