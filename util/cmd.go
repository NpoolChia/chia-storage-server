package util

import (
	"os/exec"
)

func runCmd(command string, args ...string) (result string, err error) {
	_cmd, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(_cmd, args...)
	_rcmd, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(_rcmd), nil
}
