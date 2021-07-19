BIN    := bin/
FLAGS  := -ldflags '-linkmode external -extldflags $(LDFLAGS) -w' -trimpath -buildmode=pie -mod=readonly -modcacherw
FILES  := lb lb-bash lb-diff lb-rekey lb-stats lb-pwgen lb-rw lb-totp
TARGET := $(addprefix $(BIN),$(FILES))

all: $(TARGET)

$(TARGET) : $(shell find . -type f -name "*.go") go.*
	go build -o $@ $(FLAGS) cmd/$(shell basename $@ | sed 's/lb-//g')/main.go

clean:
	rm -rf $(BIN)
	mkdir -p $(BIN)
