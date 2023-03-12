BUILD   := bin/
TARGET  := $(BUILD)lb

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	@! test -d .git || make .version | grep 'version:' | cut -d ':' -f 2 | tr '\n' '_' | sed 's/_//g' > cmd/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

.version:
	@git describe --tags --abbrev=0 | sha256sum | cut -c 1-7 | sed 's/^/version:/g'
	@git tag --points-at HEAD | grep -q '' || echo "version:-1"

unittests:
	go test -v ./...

check: $(TARGET) unittests
	cd tests && ./run.sh

clean:
	rm -rf $(BUILD)

.runci:
	rm -rf .git
	make build
	make check
