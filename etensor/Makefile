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

# tmpl is from github.com/apache/arrow/go/arrow/_tools/tmpl
# go install there to get it on your path -- not needed for
# regular builds, just if you are changing the template..
generate:
	tmpl -i -data=numeric.tmpldata numeric.gen.go.tmpl

