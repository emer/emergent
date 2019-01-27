# Basic Go makefile

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get


all: build

build: 
	$(GOBUILD) -v
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	
generate:
	tmpl -i -data=numeric.tmpldata numeric.gen.go.tmpl

