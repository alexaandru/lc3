OBJ ?= 2048

build:
	@go build .

test:
	@go test $(OPTS) ./... -cover 

bench:
	@OPTS=-bench=. make -s test

lint:
	@golangci-lint run

run: build
	@./lc3 obj/$(OBJ).obj
