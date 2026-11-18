package api

import (
	"os/exec"
)

func ptyShellCommand(shell string) *exec.Cmd {
	cmd := exec.Command(shell, "-l")
	cmd.Env = []string{
		"TERM=xterm-256color",
		"PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin",
		"LANG=en_US.UTF-8",
		"LC_ALL=en_US.UTF-8",
	}
	return cmd
}
