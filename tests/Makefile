DATA   := bin
ACTUAL := $(DATA)/actual.log

all:
	mkdir -p $(DATA)
	go run run.go 2>&1 | sed "s#$(PWD)/$(DATA)##g" | sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$$/XXXXXX/g' > $(ACTUAL)
	diff -u $(ACTUAL) expected.log
