DESTDIR ?=
PREFIX ?= /usr/local

default:

install:
	install -D -m 755 bin/tmux-rr.py $(DESTDIR)$(PREFIX)/bin/tmux-rr
	install -D -m 644 lib/tmux-rr.service $(DESTDIR)$(PREFIX)/lib/systemd/user/tmux-rr.service
