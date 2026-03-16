VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -s -w \
	-X github.com/nickw409/vex/internal/version.Version=$(VERSION) \
	-X github.com/nickw409/vex/internal/version.Commit=$(COMMIT) \
	-X github.com/nickw409/vex/internal/version.Date=$(DATE)

DIST = dist

.PHONY: build install clean test release

build:
	go build -ldflags "$(LDFLAGS)" -o vex ./cmd/vex/

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/vex/

test:
	go test ./...

release: clean
	@mkdir -p $(DIST)
	@for pair in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		os=$${pair%%/*}; arch=$${pair##*/}; \
		name=vex_$${VERSION#v}_$${os}_$${arch}; \
		echo "Building $${name}..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $(DIST)/vex ./cmd/vex/; \
		tar -czf $(DIST)/$${name}.tar.gz -C $(DIST) vex; \
		rm $(DIST)/vex; \
	done

clean:
	rm -f vex
	rm -rf $(DIST)
