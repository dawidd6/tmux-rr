DESTDIR ?=
PREFIX ?= /usr/local

default:

install:
	install -D -m 755 tmux-rr.py $(DESTDIR)$(PREFIX)/bin/tmux-rr
	install -D -m 644 tmux-rr.service $(DESTDIR)$(PREFIX)/lib/systemd/user/tmux-rr.service
