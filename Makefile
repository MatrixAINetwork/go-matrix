# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gman android ios gman-cross swarm evm all test clean
.PHONY: gman-linux gman-linux-386 gman-linux-amd64 gman-linux-mips64 gman-linux-mips64le
.PHONY: gman-linux-arm gman-linux-arm-5 gman-linux-arm-6 gman-linux-arm-7 gman-linux-arm64
.PHONY: gman-darwin gman-darwin-386 gman-darwin-amd64
.PHONY: gman-windows gman-windows-386 gman-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gman:
	build/env.sh go run build/ci.go install ./run/gman
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gman\" to launch gman."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gman.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Geth.framework\" to use the library."
clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/run/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./run/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gman-cross: gman-linux gman-darwin gman-windows gman-android gman-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gman-*

gman-linux: gman-linux-386 gman-linux-amd64 gman-linux-arm gman-linux-mips64 gman-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-*

gman-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./run/gman
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep 386

gman-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./run/gman
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep amd64

gman-linux-arm: gman-linux-arm-5 gman-linux-arm-6 gman-linux-arm-7 gman-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep arm

gman-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./run/gman
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep arm-5

gman-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./run/gman
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep arm-6

gman-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./run/gman
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep arm-7

gman-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./run/gman
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep arm64

gman-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./run/gman
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep mips

gman-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./run/gman
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep mipsle

gman-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./run/gman
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep mips64

gman-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./run/gman
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gman-linux-* | grep mips64le

gman-darwin: gman-darwin-386 gman-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gman-darwin-*

gman-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./run/gman
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gman-darwin-* | grep 386

gman-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./run/gman
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gman-darwin-* | grep amd64

gman-windows: gman-windows-386 gman-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gman-windows-*

gman-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./run/gman
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gman-windows-* | grep 386

gman-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./run/gman
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gman-windows-* | grep amd64
