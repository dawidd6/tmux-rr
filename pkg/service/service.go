package service

import (
	_ "embed"
	"os"
	"os/exec"
)

func Run() error {
	this, err := os.Executable()
	if err != nil {
		return err
	}

	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}

	return exec.Command("systemd-run", "--user", "--unit", "tmux-rr.service", "--setenv", "TMUX_PATH="+tmux, this, "daemon").Run()
}
