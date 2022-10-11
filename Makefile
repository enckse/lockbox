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

man: $(TARGET)
	help2man --include contrib/doc.sections -h help -v version ./$(TARGET) > $(BUILD)lb.man

install:
	install -Dm755 $(TARGET) $(DESTDIR)bin/lb
