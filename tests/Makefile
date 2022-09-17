BUILD := 
DATA  := bin
ALGOS := secretbox aesgcm

all: $(ALGOS)

$(ALGOS):
	./run.sh $(BUILD) $(DATA) $@
	diff -u $(DATA)/actual.log expected.log
