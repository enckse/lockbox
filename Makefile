BUILD   := bin/
TARGET  := $(BUILD)lb
VERSION ?= $(shell git log -n 1 --format=%h)

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/app/doc/*
ifeq ($(VERSION),)
	$(error version not set)
endif
	go build $(GOFLAGS) -ldflags "-X main.version=$(VERSION)" -o $@ cmd/main.go

unittests:
	go test -v ./...

check: $(TARGET) unittests
	make -C tests

clean:
	@rm -rf $(BUILD) tests/bin
	@find internal/ -type f -name "*.kdbx" -delete

.runci:
	rm -rf .git
	make build check VERSION=$(GITHUB_SHA)
