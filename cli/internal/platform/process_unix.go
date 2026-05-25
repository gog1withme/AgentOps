//go:build !windows

package platform

import (
	"os"
	"syscall"
)

func PauseProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGSTOP)
}

func KillProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}

func ShellName() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

func DetectShellType() string {
	shell := ShellName()
	if len(shell) > 0 {
		base := shell
		for i := len(shell) - 1; i >= 0; i-- {
			if shell[i] == '/' || shell[i] == '\\' {
				base = shell[i+1:]
				break
			}
		}
		switch base {
		case "zsh":
			return "zsh"
		case "bash":
			return "bash"
		case "fish":
			return "fish"
		}
	}
	return "sh"
}
