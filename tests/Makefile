BUILD := 
DATA  := bin

all:
	./run.sh $(BUILD) $(DATA)
	diff -u $(DATA)/actual.log expected.log
