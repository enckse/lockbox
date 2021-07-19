BIN := bin/
FLG := -ldflags '-linkmode external -extldflags $(LDFLAGS) -w' -trimpath -buildmode=pie -mod=readonly -modcacherw
CMD := $(shell ls cmd | sed "s|^|$(BIN)lb-|g")

all: $(CMD)

$(CMD): $(shell find . -type f -name "*.go") go.*
	go build -o $@ $(FLG) cmd/$(shell basename $@ | sed 's/lb-//g')/main.go

clean:
	rm -rf $(BIN)
	mkdir -p $(BIN)
