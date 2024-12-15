goflags := "-trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false"
target  := "target"
version := `git log -n 1 --format=%h`
object  := target / "lb"
ldflags := env_var_or_default("LDFLAGS", "")
gotest  := "LOCKBOX_CONFIG_TOML=fake go test"

default: build

build:
  mkdir -p "{{target}}"
  go build {{goflags}} -ldflags "{{ldflags}} -X main.version={{version}}" -o "{{object}}" cmd/main.go

unittest:
  {{gotest}} ./...

check: unittest tests

tests: build
  PATH="$PWD/{{target}}:$PATH" {{gotest}} cmd/main_test.go

clean:
  rm -f "{{object}}"
  find internal/ cmd/ -type f -wholename "*testdata*" -delete
  find internal/ cmd/ -type d -empty -delete
