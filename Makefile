BUILD   := bin/
TARGET  := $(BUILD)lb
TESTS   := internal/scripts/testing

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	go run internal/scripts/version/main.go cmd/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

unittest:
	go test -v ./...

check: $(TARGET) unittest
	LB_BUILD=$(PWD)/$(TARGET) make -C $(TESTS)

clean:
	rm -rf $(BUILD)
	make -C $(TESTS) clean

.runci:
	rm -rf .git
	make build
	make check
