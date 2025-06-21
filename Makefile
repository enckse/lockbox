GOFLAGS := -trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false
TARGET  := target
VERSION := $(shell git log -n 1 --format=%h)
OBJECT  := $(TARGET)/lb
GOTEST  := go test
CMD     := cmd/lb

.PHONY: $(OBJECT)

all: setup $(OBJECT)

setup:
	@test -d $(TARGET) || mkdir -p $(TARGET)

$(OBJECT):
	go build $(GOFLAGS) -ldflags "$(LDFLAGS) -X main.version=$(VERSION)" -o "$(OBJECT)" $(CMD)/main.go

unittest:
	$(GOTEST) ./...

check: unittest tests

tests: $(OBJECT)
	$(GOTEST) $(CMD)/main_test.go

clean:
	rm -f "$(OBJECT)"
	find internal/ $(CMD) -type f -wholename "*testdata*" -delete
	find internal/ $(CMD) -type d -empty -delete
