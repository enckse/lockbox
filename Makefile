VERSION := development
DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
MAIN    := $(DESTDIR)bin/lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))

.PHONY: $(TESTDIR)

all: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.*
	go build -ldflags '-X main.version=$(VERSION) -X main.mainExe=$(MAIN)' -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/main.go

$(TESTDIR):
	cd $@ && go test

check: $(TARGET) $(TESTDIR)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

install:
	install -Dm755 $(BUILD)lb $(MAIN)
