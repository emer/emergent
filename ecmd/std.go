// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecmd

import (
	"os"
)

// AddStd adds the standard command line args used by most sims
func (ar *Args) AddStd() {
	ar.AddBool("nogui", len(os.Args) > 1, "if not passing any other args and want to run nogui, use nogui")
	ar.AddBool("help", false, "show all the command line args available, then exit")
	ar.AddString("params", "", "ParamSet name to use -- must be valid name as listed in compiled-in params or loaded params")
	ar.AddString("tag", "", "extra tag to add to file names saved from this run")
	ar.AddString("note", "", "user note -- describe the run params etc")
	ar.AddInt("run", 0, "starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1")
	ar.AddInt("runs", 10, "number of runs to do (note that MaxEpcs is in paramset)")
	ar.AddInt("epochs", 150, "number of epochs per run")
	ar.AddBool("setparams", false, "if true, print a record of each parameter that is set")
	ar.AddBool("randomize", false, "If true, randomize seed for every run")
	ar.AddBool("wts", false, "if true, save final weights after each run")
	ar.AddBool("epclog", true, "if true, save train epoch log to file")
	ar.AddBool("triallog", true, "if true, save test trial log to file. May be large.")
	ar.AddBool("runlog", true, "if true, save run log to file")
	ar.AddBool("netdata", false, "if true, save network activation etc data from testing trials, for later viewing in netview")
	ar.AddString("hyperFile", "", "Name of the file to output hyperparameter data. If not empty string, program should write and then exit")
	ar.AddString("paramsFile", "", "Name of the file to input parameters from.")
}

// ProcStd processes the standard args, after Parse has been called
func (ar *Args) ProcStd() {
	if ar.Bool("help") {
		ar.Usage()
		os.Exit(0)
	}

}
