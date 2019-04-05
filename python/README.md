# Python interface to Emergent

You can now run the Go version of *emergent* via Python, using a newly-updated version of the [gopy](https://github.com/goki/gopy) tool that automatically creates Python bindings for Go packages.  Hopefully the main repository of gopy in [go-python](https://github.com/go-python/gopy) will be updated to this new version soon.

See the [GoGi Python README](https://github.com/goki/gi/blob/master/python/README.md) for more details on how the python wrapper works and how to use it for GUI-level functionality.

There is a python version of the basic leabra demo in `examples/leabra25ra/ra25.py`, which you can compare with the `ra25.go` to see how similar the python and Go code are.  The python version uses standard python data structures such as `pandas` for the input / output patterns presented to the network, and recording a log of the results, which are plotted using `matplotlib`.  Thus, it gives a good starting point for python users to build upon.

The demo shows how to live-plot the SVG output from matplotlib graphs in the gui, and also provides a GUI interface for editing Python classes, which can be used for any kind of python class object.  The functionality of this GUI will be improving rapidly over the coming months, now that the basic infrastructure is in place.

Once you edit the Network, etc, you will see the native Go editors of those objects, showing all the parameters etc, just as in the native Go version.  Thus, all these interfaces fully interoperate.

# Installation

First, you have to install the Go version of emergent: [Wiki Install](https://github.com/emer/emergent/wiki/Install).

Python version 3 (3.6 has been well tested) is recommended.

```sh
$ python3 -m pip install --upgrade pybindgen setuptools wheel pandas matplotlib
$ go get golang.org/x/tools/cmd/goimports  # gopy needs this -- you should use it too!
$ go get github.com/goki/gopy   # add -u ./... to ensure dependencies are updated
$ cd ~/go/src/github.com/goki/gopy  # use $GOPATH instead of ~/go if somewhere else
$ go install    # do go get -u ./... if this fails and try again
$ cd ~/go/src/github.com/emer/emergent/python
$ make
$ make install  # may need to do sudo make install -- installs into /usr/local/bin and python site-packages
$ cd ../examples/leabra25ra
$ pyemergent   # this was installed during make install into /usr/local/bin
$ import ra25  #this loads and runs ra25.py -- edit that and compare with ra25.go
```

* The `pyemergent` executable combines standard python and the full Go emergent and GoGi gui packages -- see the information in the GoGi python readme for more technical information about this.


