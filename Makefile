VERSION := development
DESTDIR :=
BUILD   := bin/
TARGETS := $(BUILD)lb $(BUILD)lb-totp
MAIN    := $(DESTDIR)bin/lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))

.PHONY: $(TESTDIR)

all: $(TARGETS)

$(TARGETS): cmd/**/* internal/**/*.go  go.*
	go build -ldflags '-X main.version=$(VERSION) -X main.mainExe=$(MAIN)' -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/$(shell basename $@)/main.go

$(TESTDIR):
	cd $@ && go test

check: $(TARGETS) $(TESTDIR)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

install:
	install -Dm755 $(BUILD)lb $(MAIN)
