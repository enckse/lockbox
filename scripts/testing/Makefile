BIN     := bin
ACTUAL  := $(BIN)/actual.log
RUNS    := -keyfile=true -keyfile=false
DATE    := $(shell date +%Y-%m-%d)

all: $(RUNS)

$(RUNS):
	mkdir -p $(BIN)
	rm -f $(BIN)/*.kdbx
	LB_BUILD=$(LB_BUILD) go run check.go $@ -data $(BIN) 2>&1 | sed "s#$(PWD)/$(DATA)##g" | sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$$/XXXXXX/g' | sed 's/modtime: $(DATE).*/modtime: XXXX-XX-XX/g' > $(ACTUAL)
	diff -u $(ACTUAL) expected.log

clean:
	rm -rf $(BIN)
