VERSION := development
DESTDIR :=
BUILD   := bin/
TARGETS := $(BUILD)lb $(BUILD)lb-rw $(BUILD)lb-rekey $(BUILD)lb-textconv $(BUILD)lb-totp
LIBEXEC := $(DESTDIR)libexec/lockbox/
MAIN    := $(DESTDIR)bin/lb
TESTDIR := $(shell find internal -type f -name "*test.go" -exec dirname {} \; | sort -u)

.PHONY: $(TESTDIR)

all: $(TARGETS)

$(TARGETS): cmd/**/*.go internal/*.go internal/**/*.go go.*
	go build -ldflags '-X main.version=$(VERSION) -X main.libExec=$(LIBEXEC) -X main.mainExe=$(MAIN)' -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/$(shell basename $@)/main.go

$(TESTDIR): $(TARGETS)
	cd $@ && go test

check: $(TESTDIR)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

install:
	install -Dm755 $(BUILD)lb $(MAIN)
	install -Dm755 -d $(LIBEXEC)
	install -Dm755 $(BUILD)lb-* $(LIBEXEC)
