BIN := bin/
FLG := -ldflags '-linkmode external -extldflags $(LDFLAGS) -w' -trimpath -buildmode=pie -mod=readonly -modcacherw
CMD := $(BIN)lb-diff $(BIN)lb-stats $(BIN)lb-rekey $(BIN)lb-rw $(BIN)lb $(BIN)lb-totp $(BIN)lb-pwgen

all: $(CMD)

$(CMD): $(shell find . -type f -name "*.go") go.*
	go build -o $@ $(FLG) cmd/$(shell basename $@ | sed 's/lb-//g')/main.go

clean:
	rm -rf $(BIN)
	mkdir -p $(BIN)
