goflags := "-trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false"
target  := "target"
version := `git log -n 1 --format=%h`
object  := target / "lb"
ldflags := env_var_or_default("LDFLAGS", "")

default: build

build:
  mkdir -p "{{target}}"
  go build {{goflags}} -ldflags "{{ldflags}} -X main.version={{version}}" -o "{{object}}" cmd/main.go

unittest:
  LOCKBOX_CONFIG_TOML=none go test ./...

check: unittest build
  cd tests && LOCKBOX_CONFIG_TOML=none ./run.sh

clean:
  rm -f "{{object}}"
  find internal/ tests/ -type f -wholename "*testdata*" -delete
  find internal/ tests/ -type d -empty -delete
