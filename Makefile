VERSION := development
DESTDIR :=
TARGETS := $(shell ls cmd/)
SOURCE  := $(shell find cmd/ -type f) $(shell find internal/ -type f)
LIBEXEC := $(DESTDIR)libexec/lockbox/
BUILD   := bin/

all: $(TARGETS)

$(TARGETS): $(SOURCE)
	mkdir -p $(BUILD)
	go build -ldflags '-X main.version=$(VERSION() -X main.libExec=$(LIBEXEC)' -trimpath -buildmode=pie -mod=readonly -modcacherw -o $(BUILD)$@ cmd/$@/main.go

check: $(TARGETS)
	cd tests && ./run.sh ../$(BUILD)

clean:
	rm -rf $(BUILD)
