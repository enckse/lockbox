BUILD   := bin/
TARGET  := $(BUILD)lb
VERSION ?= $(shell git log -n 1 --format=%h)
VARS    := LOCKBOX_ENV=none
DESTDIR := /usr/local/bin

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/app/doc/*
ifeq ($(VERSION),)
	$(error version not set)
endif
	go build $(GOFLAGS) -ldflags "-X main.version=$(VERSION)" -o $@ cmd/main.go

unittests:
	$(VARS) go test ./...

check: unittests runs

runs: $(TARGET)
	$(VARS) make -C tests

clean:
	@rm -rf $(BUILD) tests/bin
	@find internal/ -type f -wholename "*testdata*" -delete
	@find internal/ -type d -empty -delete

install:
	install -m755 $(TARGET) $(DESTDIR)/lb
