goflags := "-trimpath -buildmode=pie -mod=readonly -modcacherw -buildvcs=false"
target  := "target"
version := `git log -n 1 --format=%h`
object  := target / "lb"
ldflags := env_var_or_default("LDFLAGS", "")
gotest  := "LOCKBOX_CONFIG_TOML=fake go test"
files   := `find . -type f -name "*.go" | tr '\n' ' '`
cmd     := "cmd/lb"
tags    := ""

default: build

build:
  mkdir -p "{{target}}"
  go build {{goflags}} -tags={{tags}} -ldflags "{{ldflags}} -X main.version={{version}}" -o "{{object}}" {{cmd}}/main.go

unittest:
  {{gotest}} ./...

check: unittest tests

tests: features
  {{gotest}} {{cmd}}/main_test.go

features: build
  just tags=noclip object={{target}}/lb-noclip
  just tags=nototp object={{target}}/lb-nototp
  just tags=noclip,nototp object={{target}}/lb-nofeatures

clean:
  rm -f "{{object}}"
  find internal/ {{cmd}} -type f -wholename "*testdata*" -delete
  find internal/ {{cmd}} -type d -empty -delete
