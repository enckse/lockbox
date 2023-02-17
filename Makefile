DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))
DOC     := $(BUILD)doc.text
MAN     := $(BUILD)lb.man
ACTUAL  := $(BUILD)actual.log
DATE    := $(shell date +%Y-%m-%d)
RUNS    := -keyfile=true -keyfile=false

.PHONY: $(TESTDIR)

all: $(TARGET)

build: $(TARGET) $(MAN)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	./scripts/version cmd/vers.txt
	go build $(GOFLAGS) -o $@ cmd/main.go

$(TESTDIR):
	cd $@ && go test

check: $(TARGET) $(TESTDIR) $(DOC) $(RUNS)

$(RUNS):
	rm -f $(BUILD)*.kdbx
	LB_BUILD=$(TARGET) TEST_DATA=$(BUILD) SCRIPTS=$(PWD)/scripts/ go run scripts/check.go $@ 2>&1 | sed "s#$(PWD)/$(DATA)##g" | sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$$/XXXXXX/g' | sed 's/modtime: $(DATE).*/modtime: XXXX-XX-XX/g' > $(ACTUAL)
	diff -u $(ACTUAL) scripts/tests.expected.log

clean:
	rm -rf $(BUILD)

$(DOC): $(TARGET)
	$(TARGET) help -verbose > $(DOC)
	test -s $(DOC)

$(MAN): $(TARGET) $(DOC)
	help2man --include $(DOC) -h help -v version -o $(MAN) ./$(TARGET)

install:
	install -Dm644 $(MAN) $(DESTDIR)share/man/man1/lb.1
	install -Dm755 $(TARGET) $(DESTDIR)bin/lb

