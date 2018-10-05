package deploy

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func ExecuteCommand(command string, timeout int) (context.CancelFunc, chan (bool), error) {
	var ctx context.Context
	var cancel context.CancelFunc

	toks := strings.Split(command, " ")
	name := toks[0]
	args := toks[1:]

	if timeout == 0 {
		ctx, cancel = context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		copyStderr(cmd)
		copyStdout(cmd)

		if err := cmd.Start(); err != nil {
			return nil, nil, err
		}

		killed := make(chan (bool), 1)

		go func() {
			<-ctx.Done()

			log.Printf("Goroutine caught signal for PID %d\n", -cmd.Process.Pid)

			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				log.Println(err)
				killed <- false
			}

			killed <- true
		}()

		return cancel, killed, nil
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
		defer cancel()

		cmd := exec.CommandContext(ctx, name, args...)
		copyStderr(cmd)
		copyStdout(cmd)

		if err := cmd.Start(); err != nil {
			log.Println("Failed starting command")
		}

		err := cmd.Wait()

		if ctx.Err() == context.DeadlineExceeded {
			return nil, nil, ctx.Err()
		} else if err != nil {
			return nil, nil, err
		}

		return nil, nil, nil
	}
}

func copyStdout(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Failed capturing command stdout")
	}

	go io.Copy(os.Stdout, stdout)
}

func copyStderr(cmd *exec.Cmd) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("Failed capturing command stderr")
	}

	go io.Copy(os.Stderr, stderr)
}
