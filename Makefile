BUILD   := bin/
TARGET  := $(BUILD)lb

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	./scripts/version/configure cmd/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

unittest:
	go test -v ./...

check: $(TARGET) unittest
	make -C scripts/testing LB_BUILD=$(PWD)/$(TARGET) TEST_DATA=$(PWD)/$(BUILD) SCRIPTS=$(PWD)/scripts/testing/

clean:
	rm -rf $(BUILD)
