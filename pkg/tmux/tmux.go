package tmux

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
)

var ErrServerAlreadyStarted error = errors.New("tmux server is already started")

type Session struct {
	Name    string    `json:"name" tmux:"session_name"`
	Windows []*Window `json:"windows,omitempty"`
}

type Window struct {
	Name   string  `json:"name" tmux:"window_name"`
	Index  string  `json:"index" tmux:"window_index"`
	Layout string  `json:"layout" tmux:"window_layout"`
	Width  string  `json:"width" tmux:"window_width"`
	Height string  `json:"height" tmux:"window_height"`
	Panes  []*Pane `json:"panes,omitempty"`
}

type Pane struct {
	Index      string `json:"index" tmux:"pane_index"`
	Path       string `json:"path" tmux:"pane_current_path"`
	Command    string `json:"command" tmux:"@tmux-rr-command"`
	Scrollback string `json:"scrollback"`
	ID         string `json:"id" tmux:"pane_id"`
	TTY        string `json:"tty" tmux:"pane_tty"`
}

func marshalFormat[T any]() (string, error) {
	typ := reflect.TypeFor[T]()
	object := reflect.New(typ)
	value := object.Elem()

	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("tmux")
		if tag != "" {
			value.Field(i).SetString("#{" + tag + "}")
		}
	}

	format, err := json.Marshal(object.Interface())
	if err != nil {
		return "", err
	}

	return string(format), nil
}

func unmarshalFormattedOutput[T any](data []byte) ([]*T, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	objects := make([]*T, 0)

	for {
		object := new(T)
		err := decoder.Decode(object)
		if errors.Is(err, io.EOF) {
			return objects, nil
		} else if err != nil {
			return nil, err
		}

		objects = append(objects, object)
	}
}

func tmuxCommand(args ...string) *exec.Cmd {
	tmuxPath := os.Getenv("TMUX_PATH")
	if tmuxPath == "" {
		tmuxPath = "tmux"
	}

	tmuxSocket := os.Getenv("TMUX_SOCKET")
	if tmuxSocket != "" {
		args = append([]string{"-L", tmuxSocket}, args...)
	}

	return exec.Command(tmuxPath, args...)
}

func run(args ...string) ([]byte, error) {
	cmd := tmuxCommand(args...)
	out, err := cmd.CombinedOutput()
	out = bytes.TrimSpace(out)
	if err != nil {
		return nil, fmt.Errorf("cmd: %v\nout: %v\nerr: %w", cmd, string(out), err)
	}

	runFormatted[Session](func(s string) []string {
		return []string{"abc"}
	})

	return out, nil
}

func runFormatted[T any](f func(string) []string) ([]*T, error) {
	format, err := marshalFormat[T]()
	if err != nil {
		return nil, err
	}

	output, err := run(f(format)...)
	if err != nil {
		return nil, err
	}

	return unmarshalFormattedOutput[T](output)
}

func HasSession(session *Session) (bool, error) {
	format, err := marshalFormat[Session]()
	if err != nil {
		return false, err
	}

	output, err := run("list-sessions", "-F", format)
	if err != nil {
		return false, err
	}

	existingSessions, err := unmarshalFormattedOutput[Session](output)
	if err != nil {
		return false, err
	}

	for _, existingSession := range existingSessions {
		if existingSession.Name == session.Name {
			return true, nil
		}
	}

	return false, nil
}

func NewSession(session *Session, window *Window, pane *Pane) error {
	format, err := marshalFormat[Pane]()
	if err != nil {
		return err
	}

	output, err := run("new-session", "-dPF", format, "-x", window.Width, "-y", window.Height, "-s", session.Name, "-n", window.Name, "-c", pane.Path, "sleep", "infinity")
	if err != nil {
		return err
	}

	newPane, err := unmarshalFormattedOutput[Pane](output)
	if err != nil {
		return err
	}

	pane.ID = newPane[0].ID
	pane.TTY = newPane[0].TTY

	return nil
}

func NewWindow(session *Session, window *Window, pane *Pane) error {
	target := fmt.Sprintf("=%s", session.Name)

	format, err := marshalFormat[Pane]()
	if err != nil {
		return err
	}

	output, err := run("new-window", "-dPF", format, "-t", target, "-n", window.Name, "-c", pane.Path, "sleep", "infinity")
	if err != nil {
		return err
	}

	newPane, err := unmarshalFormattedOutput[Pane](output)
	if err != nil {
		return err
	}

	pane.ID = newPane[0].ID
	pane.TTY = newPane[0].TTY

	return nil
}

func SplitWindow(session *Session, pane *Pane) error {
	target := fmt.Sprintf("=%s:{end}", session.Name)

	format, err := marshalFormat[Pane]()
	if err != nil {
		return err
	}

	output, err := run("split-window", "-dPF", format, "-t", target, "-c", pane.Path, "sleep", "infinity")
	if err != nil {
		return err
	}

	newPane, err := unmarshalFormattedOutput[Pane](output)
	if err != nil {
		return err
	}

	pane.ID = newPane[0].ID
	pane.TTY = newPane[0].TTY

	return nil
}

func RespawnPane(pane *Pane) error {
	format, err := marshalFormat[Pane]()
	if err != nil {
		return err
	}

	// TODO: swap $SHELL with default-command from tmux
	output, err := run("respawn-pane", "-k", "-t", pane.ID, "-c", pane.Path, "sh", "-c", "tmux wait-for tmux-rr; exec \"$SHELL\"", ";", "display-message", "-p", "-t", pane.ID, "-F", format)

	newPane, err := unmarshalFormattedOutput[Pane](output)
	if err != nil {
		return err
	}

	pane.ID = newPane[0].ID
	pane.TTY = newPane[0].TTY

	return nil
}

func ResizePane(pane *Pane) error {
	_, err := run("resize-pane", "-t", pane.ID, "-U", "999")
	return err
}

func WaitForS() error {
	_, err := run("wait-for", "-S", "tmux-rr")
	return err
}

func SelectLayout(session *Session, window *Window) error {
	target := fmt.Sprintf("=%s:{end}", session.Name)
	_, err := run("select-layout", "-t", target, window.Layout)
	return err
}

func ListSessions() ([]*Session, error) {
	format, err := marshalFormat[Session]()
	if err != nil {
		return nil, err
	}

	output, err := run("list-sessions", "-F", format)
	if err != nil {
		return nil, err
	}

	return unmarshalFormattedOutput[Session](output)
}

func ListWindows(session *Session) ([]*Window, error) {
	target := fmt.Sprintf("=%s", session.Name)

	format, err := marshalFormat[Window]()
	if err != nil {
		return nil, err
	}

	output, err := run("list-windows", "-F", format, "-t", target)
	if err != nil {
		return nil, err
	}

	return unmarshalFormattedOutput[Window](output)
}

func ListPanes(session *Session, window *Window) ([]*Pane, error) {
	target := fmt.Sprintf("=%s:%s", session.Name, window.Index)

	format, err := marshalFormat[Pane]()
	if err != nil {
		return nil, err
	}

	output, err := run("list-panes", "-F", format, "-t", target)
	if err != nil {
		return nil, err
	}

	return unmarshalFormattedOutput[Pane](output)
}

func CapturePane(session *Session, window *Window, pane *Pane) (string, error) {
	target := fmt.Sprintf("=%s:%s.%s", session.Name, window.Index, pane.Index)
	data, err := run("capture-pane", "-epJS", "0", "-t", target)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(append(data, '\n')), nil
}

func WriteTTY(pane *Pane) error {
	tty, err := os.OpenFile(pane.TTY, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer tty.Close()

	scrollback, err := base64.StdEncoding.DecodeString(pane.Scrollback)
	if err != nil {
		return err
	}

	if _, err = tty.Write(scrollback); err != nil {
		return err
	}

	return nil
}

func SendCommand(pane *Pane) error {
	if pane.Command == "" {
		return nil
	}

	command, err := base64.StdEncoding.DecodeString(pane.Command)
	if err != nil {
		return err
	}

	_, err = run("send-keys", "-t", pane.ID, "-l", "--", string(command))
	return err
}

func AutomaticRename(session *Session, window *Window) error {
	output, err := run("show-options", "-gv", "automatic-rename")
	if err != nil {
		return err
	}

	output, err = run("set-options", "-w", "automatic-rename", string(output))
	if err != nil {
		return err
	}

	return nil
}

func StartServer() error {
	if _, err := run("server-info"); err == nil {
		return ErrServerAlreadyStarted
	}

	_, err := run("start-server", ";", "set-option", "-s", "exit-empty", "off")
	return err
}

func KillServer() error {
	_, err := run("kill-server")
	return err
}
