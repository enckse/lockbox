DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))

.PHONY: $(TESTDIR)

all: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.*
	./contrib/version
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/main.go

$(TESTDIR):
	cd $@ && go test

check: $(TARGET) $(TESTDIR)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

install:
	install -Dm755 $(BUILD)lb $(DESTDIR)bin/lb
