GOPATH := $(shell go env GOPATH)

build:
	@go generate && wails build --clean
	
dev:
	@go generate && wails build --clean && ./build/bin/smDesktop -dev
	
run:
	@go generate && wails build --clean && ./build/bin/smDesktop

install:
	@go generate && wails build --clean && cp ./build/bin/smDesktop $(GOPATH)/bin/
