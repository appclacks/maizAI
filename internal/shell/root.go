package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type ShellRunner struct {
}

func (s *ShellRunner) Execute(ctx context.Context, command string, arguments ...string) (string, error) {
	var stdErr, stdOut bytes.Buffer
	cmd := exec.CommandContext(ctx, command, arguments...)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	if err := cmd.Run(); err != nil {
		var errorMsg string
		exitErr, isExitError := err.(*exec.ExitError)
		if isExitError {
			errorMsg = fmt.Sprintf("The command failed with code=%d, stderr=%s", exitErr.ExitCode(), stdErr.String())
		} else {
			errorMsg = fmt.Sprintf("The command failed, stderr=%s", stdErr.String())
		}
		return "", fmt.Errorf("%s, %w", errorMsg, err)
	}

	return stdOut.String(), nil
}
