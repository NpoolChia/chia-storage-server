package util

import (
	"os/exec"
	"syscall"
)

func runCmd(command string, args ...string) (result int, err error) {
	_cmd, err := exec.LookPath(command)
	if err != nil {
		return 1, err
	}
	err = exec.Command(_cmd, args...).Run()
	if err != nil {
		code, ok := err.(*exec.ExitError)
		if ok {
			ws := code.Sys().(syscall.WaitStatus)
			if ws.ExitStatus() == 0 {
				return 0, nil
			}
			return 1, err
		}
		return 1, err
	}
	return 0, nil
}
