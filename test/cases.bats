#!/usr/bin/env bash

# Variables

IMAGE_NAME=tmux-rr-test
CONTAINER_NAME=tmux-rr-test
CONTAINER_USER_ID=1000
CONTAINER_USER_NAME=ubuntu
CONTAINER_XDG_RUNTIME_DIR="/run/user/$CONTAINER_USER_ID"
STATE_FILE_PATH="/home/test/.local/state/tmux-rr.json"

# Helpers

function debug_shell() {
    echo 'DROPPING INTO INTERACTIVE DEBUG SHELL!' >/dev/tty
    echo 'Press CTRL-D to exit or write "exit" command.' >/dev/tty
    "$SHELL" </dev/tty >/dev/tty 2>&1
}

function debug_shell_container() {
    echo 'DROPPING INTO INTERACTIVE DEBUG SHELL!' >/dev/tty
    echo 'Press CTRL-D to exit or write "exit" command.' >/dev/tty
    shell_container "$1" </dev/tty >/dev/tty 2>&1
}

function build_image() {
    podman build -t "$IMAGE_NAME" -f test/Containerfile test
}

function run_container() {
    podman run -dt --name "$CONTAINER_NAME" -v .:/wd -w /wd "$IMAGE_NAME"
}

function restart_container() {
    podman restart "$CONTAINER_NAME"
}

function destroy_container() {
    podman rm -f "$CONTAINER_NAME"
}

function shell_container() {
    podman exec -it -u "$CONTAINER_USER_ID" -e XDG_RUNTIME_DIR="$CONTAINER_XDG_RUNTIME_DIR" "$CONTAINER_NAME" "$@"
}

function exec_container() {
    podman exec -u "$CONTAINER_USER_ID" -e XDG_RUNTIME_DIR="$CONTAINER_XDG_RUNTIME_DIR" "$CONTAINER_NAME" "$@"
}

function wait_until_succeeds() {
    local max=50
    local current=0
    while ((current < max)); do
        if "$@"; then
            return 0
        fi
        sleep 0.1
        current=$((current + 1))
    done
    return 1
}

function wait_for_container() {
    wait_until_succeeds exec_container systemctl --user is-system-running --wait
}

function install_project() {
    case "$1" in
    bash)
        # TODO: impl
        # shellcheck disable=SC2016
        # exec_container bash -c 'eval "$(./tmux-rr init bash)"'
        :
        ;;
    fish)
        exec_container fish -c './tmux-rr init fish | source'
        ;;
    esac
    exec_container systemctl --user daemon-reload
    exec_container systemctl --user start tmux-rr
}

function restart_project() {
    exec_container systemctl --user restart tmux-rr
}

function list_panes {
    run exec_container tmux -N list-panes -a -F '
    #{session_name}
    #{window_index}
    #{window_name}
    #{window_panes}
    #{window_layout}
    #{window_width}
    #{window_height}
    #{pane_index}
    #{pane_current_path}'
    # shellcheck disable=SC2154
    [ "$status" -eq 0 ]
    [ -n "$output" ]
    echo "$output"
}

# Setups

function setup_file() {
    build_image
}

function setup() {
    run_container
    wait_for_container
}

# Teardowns

function teardown() {
    destroy_container
}

# Cases

@test "fish_test" {
    install_project fish
    exec_container tmux -N set-option -g automatic-rename off
    exec_container tmux -N new-session -d -s test-session
    exec_container tmux -N new-window -d -t =test-session -c /tmp sleep inf
    expected="$(list_panes)"
    restart_project
    actual="$(list_panes)"
    diff -u <(echo "$expected") <(echo "$actual")
}
