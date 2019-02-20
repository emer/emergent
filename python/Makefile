# Makefile for gopy pkg generation of python bindings to github.com/emer/emergent

all: gen

gen:
	gopy pkg -vm=python3 -exclude="python" github.com/emer/emergent

install:
	# this does a local install of the package, building the sdist and then directly installing it
	rm -rf dist build */*.egg-info *.egg-info
	python3 setup.py sdist
	pip install dist/*.tar.gz

