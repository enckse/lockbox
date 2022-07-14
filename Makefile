VERSION := development
DESTDIR :=
BUILD   := bin/
TARGETS := $(BUILD)lb $(BUILD)lb-rw $(BUILD)lb-bash $(BUILD)lb-rekey $(BUILD)lb-diff $(BUILD)lb-totp
LIBEXEC := $(DESTDIR)libexec/lockbox/

all: $(TARGETS)

$(TARGETS): cmd/$@/* internal/* go.*
	go build -ldflags '-X main.version=$(VERSION) -X main.libExec=$(LIBEXEC)' -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/$(shell basename $@)/main.go

check: $(TARGETS)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

install:
	install -Dm755 $(BUILD)lb $(DESTDIR)bin/lb
	install -Dm755 -d $(LIBEXEC)
	install -Dm755 $(BUILD)lb-* $(LIBEXEC)
