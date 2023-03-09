$(RUNS):
	mkdir -p $(BIN)
	rm -f $(BIN)/*.kdbx
	LB_BUILD=$(LB_BUILD) go run check.go $@ -run=true -data $(BIN) > $(ACTUAL) 2>&1
	go run check.go -run=false -data $(ACTUAL)
	diff -u $(ACTUAL) expected.log
