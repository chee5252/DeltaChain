.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/deltachaineum/go-deltachaineum/cmd/swarm
	go build -o bin/gdch     github.com/deltachaineum/go-deltachaineum/cmd/gdch
	go build -o bin/bootnode github.com/deltachaineum/go-deltachaineum/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/deltachaineum/go-deltachaineum/swarm/...
