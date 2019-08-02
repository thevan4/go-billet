package executor

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
)

// Incoming data must be an array of bytes without any encoding

// Execute command with arguments
func Execute(fullCommand, contextFolder string, arguments []string) ([]byte, []byte, int, error) {
	var emptyBytes []byte
	cmd := exec.Command(fullCommand, arguments...)
	cmd.Dir = contextFolder

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return emptyBytes, emptyBytes, 1, fmt.Errorf("can't get command out pipe, got error: %v", err)
	}
	defer stdoutPipe.Close()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return emptyBytes, emptyBytes, 1, fmt.Errorf("can't get command err pipe, got error: %v", err)
	}
	defer stderrPipe.Close()

	if err := cmd.Start(); err != nil {
		return emptyBytes, emptyBytes, 1, fmt.Errorf("can't run command, got error: %v", err)
	}

	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return emptyBytes, emptyBytes, 1, fmt.Errorf("can't read stdout, got error: %v", err)
	}

	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return emptyBytes, emptyBytes, 1, fmt.Errorf("can't read stdout, got error: %v", err)
	}

	var exitCode int
	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	return stdout, stderr, exitCode, nil
}
