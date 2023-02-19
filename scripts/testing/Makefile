ACTUAL  := $(TEST_DATA)actual.log
RUNS    := -keyfile=true -keyfile=false
DATE    := $(date +%Y-%m-%d)

all: $(RUNS)

$(RUNS):
	rm -f $(TEST_DATA)*.kdbx
	LB_BUILD=$(LB_BUILD) TEST_DATA=$(TEST_DATA) SCRIPTS=$(SCRIPTS) go run check.go $@ 2>&1 | sed "s#$(PWD)/$(DATA)##g" | sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$$/XXXXXX/g' | sed 's/modtime: $(DATE).*/modtime: XXXX-XX-XX/g' > $(ACTUAL)
	diff -u $(ACTUAL) expected.log
