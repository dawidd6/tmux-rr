package state

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/dawidd6/tmux-rr/pkg/tmux"
)

type State struct {
	Sessions []*tmux.Session `json:"sessions"`
}

func file() string {
	return path.Join(xdg.StateHome, "tmux-rr.json")
}

func read() (*State, error) {
	file := file()
	state := &State{}

	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	save, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(save, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func write(state *State) error {
	file := file()

	err := os.MkdirAll(path.Dir(file), 0o750)
	if err != nil {
		return err
	}

	save, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, save, 0o600)
}

func Clear() error {
	file := file()

	err := os.Remove(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func Save() error {
	state := &State{}
	err := error(nil)

	state.Sessions, err = tmux.ListSessions()
	if err != nil {
		return err
	}

	for _, session := range state.Sessions {
		session.Windows, err = tmux.ListWindows(session)
		if err != nil {
			return err
		}

		for _, window := range session.Windows {
			window.Panes, err = tmux.ListPanes(session, window)
			if err != nil {
				return err
			}

			for _, pane := range window.Panes {
				pane.Scrollback, err = tmux.CapturePane(session, window, pane)
				if err != nil {
					return err
				}
			}
		}
	}

	return write(state)
}

func Restore() error {
	state, err := read()
	if err != nil {
		return err
	}

	if state == nil {
		return nil
	}

	for _, session := range state.Sessions {
		has, err := tmux.HasSession(session)
		if err != nil {
			return err
		}

		for _, window := range session.Windows {
			for i, pane := range window.Panes {
				if i == 0 && !has {
					err = tmux.NewSession(session, window, pane)
					if err != nil {
						return err
					}
				}

				if i == 0 && has {
					err = tmux.NewWindow(session, window, pane)
					if err != nil {
						return err
					}
				}

				if i > 0 {
					err = tmux.SplitWindow(session, pane)
					if err != nil {
						return err
					}

					err = tmux.ResizePane(pane)
					if err != nil {
						return err
					}
				}
			}

			err = tmux.SelectLayout(session, window)
			if err != nil {
				return err
			}

			for _, pane := range window.Panes {
				err = tmux.RespawnPane(pane)
				if err != nil {
					return err
				}

				err = tmux.WriteTTY(pane)
				if err != nil {
					return err
				}

				err = tmux.WaitForS()
				if err != nil {
					return err
				}

				err = tmux.SendCommand(pane)
				if err != nil {
					return err
				}
			}

			err = tmux.AutomaticRename(session, window)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
