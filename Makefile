DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))
DOC     := $(BUILD)doc.text
MAN     := $(BUILD)lb.man
DOCTEXT := scripts/doc.sections

.PHONY: $(TESTDIR)

all: $(TARGET)

build: $(TARGET) $(MAN)

$(TARGET): cmd/main.go internal/**/*.go  go.* internal/cli/completions*
	./scripts/version
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -o $@ cmd/main.go

$(TESTDIR):
	cd $@ && go test

check: $(TARGET) $(TESTDIR)
	cd tests && make BUILD=../$(BUILD)

clean:
	rm -rf $(BUILD)

$(DOC): $(TARGET) $(DOCTEXT)
	@cat $(DOCTEXT) > $(DOC)
	$(TARGET) env -defaults >> $(DOC)

$(MAN): $(TARGET) $(DOC)
	help2man --include $(DOC) -h help -v version -o $(MAN) ./$(TARGET)

install:
	install -Dm644 $(MAN) $(DESTDIR)share/man/man1/lb.1
	install -Dm755 $(TARGET) $(DESTDIR)bin/lb
