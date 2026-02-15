GOFLAGS := -trimpath -mod=readonly -modcacherw -buildvcs=false
TARGET  := dist
VERSION ?= "$(shell git describe --abbrev=0 --tags)-$(shell git log -n 1 --format=%h)"
OBJECT  := $(TARGET)/lb
GOTEST  := go test
CMD     := cmd/lb

.PHONY: $(OBJECT)

all: setup $(OBJECT)

setup:
	@test -d $(TARGET) || mkdir -p $(TARGET)

generate:
	go generate ./...

$(OBJECT):
	go build $(GOFLAGS) -ldflags "$(LDFLAGS) -X main.version=$(VERSION)" -o "$(OBJECT)" $(CMD)/main.go

unittest:
	$(GOTEST) ./...

check: generate unittest tests

tests: $(OBJECT)
	$(GOTEST) $(CMD)/main_test.go

clean:
	rm -f "$(OBJECT)"*
	find internal/ $(CMD) -type f -wholename "*testdata*" -delete
	find internal/ $(CMD) -type d -empty -delete

_release:
	mv $(OBJECT) $(OBJECT)-$(GOOS)-$(GOARCH)

releases: clean $(OBJECT)
