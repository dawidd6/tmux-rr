package shell

import (
	"embed"
	"fmt"
	"path"
)

//go:embed scripts
var scripts embed.FS

type Shell = string

const (
	Bash Shell = "bash"
	Fish Shell = "fish"
	Zsh  Shell = "zsh"
)

func Supported() []Shell {
	return []Shell{Bash, Fish, Zsh}
}

func Print(shell Shell) error {
	script, err := scripts.ReadFile(path.Join("scripts", "shell."+shell))
	if err != nil {
		return err
	}

	fmt.Println(string(script))
	return nil
}
