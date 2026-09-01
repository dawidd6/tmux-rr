# @fish-lsp-disable-next-line 4004
function __tmux_rr_command_saver_preexec --on-event fish_preexec
    test -n "$TMUX" && tmux -N set-option -pq @tmux-rr-command (printf '%s' "$argv" | base64 -w0)
end
# @fish-lsp-disable-next-line 4004
function __tmux_rr_command_saver_postexec --on-event fish_postexec
    test -n "$TMUX" && tmux -N set-option -pqu @tmux-rr-command
end
