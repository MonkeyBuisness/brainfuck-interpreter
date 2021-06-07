all: run

run:
	go run main.go --if ./examples/factorial.bf --of ./out/output.txt

shell:
	go run main.go
