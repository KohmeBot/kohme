package util

import (
	"os"
	"os/exec"
)

func Restart() (exit func(), err error) {
	exe, err := os.Executable()
	if err != nil {
		return func() {
			os.Exit(1)
		}, err
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return func() {
		os.Exit(0)
	}, cmd.Start()
}
