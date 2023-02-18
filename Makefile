DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
DOC     := $(BUILD)doc.text
ACTUAL  := $(BUILD)actual.log
DATE    := $(date +%Y-%m-%d)
RUNS    := -keyfile=true -keyfile=false

all: $(TARGET)

build: $(TARGET)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	./scripts/version cmd/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

unittest:
	go test -v ./...

check: $(TARGET) unittest $(RUNS)
	$(TARGET) help -verbose > $(DOC)
	test -s $(DOC)

$(RUNS):
	rm -f $(BUILD)*.kdbx
	LB_BUILD=$(TARGET) TEST_DATA=$(BUILD) SCRIPTS=$(PWD)/scripts/ go run scripts/check.go $@ 2>&1 | sed "s#$(PWD)/$(DATA)##g" | sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$$/XXXXXX/g' | sed 's/modtime: $(DATE).*/modtime: XXXX-XX-XX/g' > $(ACTUAL)
	diff -u $(ACTUAL) scripts/tests.expected.log

clean:
	rm -rf $(BUILD)
