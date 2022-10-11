DESTDIR :=
BUILD   := bin/
TARGET  := $(BUILD)lb
TESTDIR := $(sort $(dir $(wildcard internal/**/*_test.go)))
DOC     := contrib/doc.sections
MAN     := $(BUILD)lb.man

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

$(MAN): $(TARGET) $(DOC)
	help2man --include $(DOC) -h help -v version ./$(TARGET) > $(MAN)

install:
	install -Dm755 $(TARGET) $(DESTDIR)bin/lb
