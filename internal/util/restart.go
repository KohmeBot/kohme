package util

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func Restart() (exit func(), err error) {
	if runtime.GOOS == "windows" {
		return func() {
			os.Exit(1)
		}, fmt.Errorf("unsupport windows")
	}

	exe, err := os.Executable()
	if err != nil {
		return func() {
			os.Exit(1)
		}, err
	}

	return func() {
		os.Exit(0)
	}, syscall.Exec(exe, os.Args, os.Environ())
}
