BUILD   := bin/
TARGET  := $(BUILD)lb
TESTS   := scripts/testing

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	./scripts/version/configure internal/app/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

unittest:
	go test -v ./...

check: $(TARGET) unittest
	LB_BUILD=$(PWD)/$(TARGET) make -C $(TESTS)

clean:
	rm -rf $(BUILD)
	make -C $(TESTS) clean
