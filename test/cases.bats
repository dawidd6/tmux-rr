#!/usr/bin/env bash

# Variables

IMAGE_NAME=tmux-rr-test
CONTAINER_NAME=tmux-rr-test
CONTAINER_USER_ID=1000
CONTAINER_XDG_RUNTIME_DIR="/run/user/$CONTAINER_USER_ID"
STATE_FILE_PATH="/home/test/.local/state/tmux-rr.json"

# Helpers

function debug_shell() {
    echo 'DROPPING INTO INTERACTIVE DEBUG SHELL!' >/dev/tty
    echo 'Press CTRL-D to exit or write "exit" command.' >/dev/tty
    "$SHELL" </dev/tty >/dev/tty 2>&1
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

function do_in_container() {
    podman exec -u "$CONTAINER_USER_ID" -e XDG_RUNTIME_DIR="$CONTAINER_XDG_RUNTIME_DIR" "$CONTAINER_NAME" "$@"
}

function simulate_reboot() {
    restart_container
    wait_for_container
    wait_for_project
}

function wait_until_succeeds() {
    local max=10
    local current=0
    while ((current < max)); do
        if "$@"; then
            break
        fi
        sleep 1s
        current=$((current + 1))
    done
}

function wait_for_container() {
    wait_until_succeeds do_in_container sudo systemctl is-system-running --wait
}

function wait_for_project() {
    wait_until_succeeds do_in_container systemctl --user is-active tmux-rr
}

function install_project() {
    do_in_container sudo make install
    do_in_container systemctl --user enable --now tmux-rr
}

function list_panes {
    run do_in_container tmux -N list-panes -a -F '
    #{session_name}
    #{window_name}
    #{window_layout}'
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
    install_project
    wait_for_project
}

# Teardowns

function teardown() {
    destroy_container
}

# Cases

function main_test {
    # debug_shell
    do_in_container tmux -N new-session -d -s test123
    expected="$(list_panes)"
    simulate_reboot
    actual="$(list_panes)"
    [ "$expected" = "$actual" ]
}

if [ -n "$CI" ]; then
    for i in $(seq 1 10); do
        bats_test_function -- main_test "$i"
    done
else
    bats_test_function -- main_test
fi
