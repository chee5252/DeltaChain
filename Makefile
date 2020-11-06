# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gdch android ios gdch-cross swarm evm all test clean
.PHONY: gdch-linux gdch-linux-386 gdch-linux-amd64 gdch-linux-mips64 gdch-linux-mips64le
.PHONY: gdch-linux-arm gdch-linux-arm-5 gdch-linux-arm-6 gdch-linux-arm-7 gdch-linux-arm64
.PHONY: gdch-darwin gdch-darwin-386 gdch-darwin-amd64
.PHONY: gdch-windows gdch-windows-386 gdch-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gdch:
	build/env.sh go run build/ci.go install ./cmd/gdch
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gdch\" to launch gdch."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gdch.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gdch.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	./build/clean_go_build_cache.sh
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

swarm-devtools:
	env GOBIN= go install ./cmd/swarm/mimegen

# Cross Compilation Targets (xgo)

gdch-cross: gdch-linux gdch-darwin gdch-windows gdch-android gdch-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gdch-*

gdch-linux: gdch-linux-386 gdch-linux-amd64 gdch-linux-arm gdch-linux-mips64 gdch-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-*

gdch-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gdch
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep 386

gdch-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gdch
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep amd64

gdch-linux-arm: gdch-linux-arm-5 gdch-linux-arm-6 gdch-linux-arm-7 gdch-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep arm

gdch-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gdch
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep arm-5

gdch-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gdch
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep arm-6

gdch-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gdch
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep arm-7

gdch-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gdch
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep arm64

gdch-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gdch
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep mips

gdch-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gdch
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep mipsle

gdch-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gdch
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep mips64

gdch-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gdch
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gdch-linux-* | grep mips64le

gdch-darwin: gdch-darwin-386 gdch-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gdch-darwin-*

gdch-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gdch
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-darwin-* | grep 386

gdch-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gdch
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-darwin-* | grep amd64

gdch-windows: gdch-windows-386 gdch-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gdch-windows-*

gdch-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gdch
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-windows-* | grep 386

gdch-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gdch
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gdch-windows-* | grep amd64
