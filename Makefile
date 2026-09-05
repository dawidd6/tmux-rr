.PHONY: build
build:
	go build ./cmd/tmux-rr

.PHONY: unit
unit:
	go test -v ./...

.PHONY: image
image: build
	podman build -t tmux-rr .

.PHONY: container
container: image
	podman run --cap-add SYS_ADMIN -it --rm -v .:/wd -w /wd tmux-rr

.PHONY: e2e
e2e: image
	robot --loglevel TRACE --outputdir e2e/results -- e2e/tests.robot

.PHONY: ci
ci: unit e2e
