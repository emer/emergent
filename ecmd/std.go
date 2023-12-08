// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecmd

import (
	"os"

	"github.com/emer/emergent/v2/elog"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/empi/v2/mpi"
)

// AddStd adds the standard command line args used by most sims
func (ar *Args) AddStd() {
	ar.AddBool("nogui", len(os.Args) > 1, "if not passing any other args and want to run nogui, use nogui")
	ar.AddBool("help", false, "show all the command line args available, then exit")
	ar.AddString("params", "", "ParamSet name to use -- must be valid name as listed in compiled-in params or loaded params")
	ar.AddString("tag", "", "extra tag to add to file names and logs saved from this run")
	ar.AddString("note", "", "user note -- describe the run params etc")
	ar.AddInt("run", 0, "starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1")
	ar.AddInt("runs", 10, "number of runs to do (note that MaxEpcs is in paramset)")
	ar.AddInt("epochs", 150, "number of epochs per run")
	ar.AddBool("setparams", false, "if true, print a record of each parameter that is set")
	ar.AddBool("randomize", false, "If true, randomize seed for every run")
	ar.AddBool("wts", false, "if true, save final weights after each run")
	ar.AddBool("epclog", true, "if true, save train epoch log to file")
	ar.AddBool("triallog", false, "if true, save train trial log to file. May be large.")
	ar.AddBool("runlog", true, "if true, save run log to file")
	ar.AddBool("tstepclog", false, "if true, save testing epoch log to file")
	ar.AddBool("tsttriallog", false, "if true, save testing trial log to file. May be large.")
	ar.AddBool("netdata", false, "if true, save network activation etc data from testing trials, for later viewing in netview")
	ar.AddString("hyperFile", "", "Name of the file to output hyperparameter data. If not empty string, program should write and then exit")
	ar.AddString("paramsFile", "", "Name of the file to input parameters from.")
	ar.AddBool("gpu", false, "Use the GPU to run the model -- typically faster for larger models.")
}

// LogFileName returns a standard log file name as netName_runName_logName.tsv
func LogFileName(logName, netName, runName string) string {
	return netName + "_" + runName + "_" + logName + ".tsv"
}

// ProcStd processes the standard args, after Parse has been called
// for help, note, params, tag and wts
func (ar *Args) ProcStd(params *emer.Params) {
	if ar.Bool("help") {
		ar.Usage()
		os.Exit(0)
	}
	if note := ar.String("note"); note != "" {
		mpi.Printf("note: %s\n", note)
	}
	if pars := ar.String("params"); pars != "" {
		params.ExtraSets = pars
		mpi.Printf("Using ParamSet: %s\n", params.ExtraSets)
	}
	if tag := ar.String("tag"); tag != "" {
		params.Tag = tag
	}
	if ar.Bool("wts") {
		mpi.Printf("Saving final weights per run\n")
	}

}

// ProcStdLogs processes the standard args for log files,
// setting the log files for standard log file names using netName
// and params.RunName to identify the network / sim and run params, tag,
// and starting run number
func (ar *Args) ProcStdLogs(logs *elog.Logs, params *emer.Params, netName string) {
	runName := params.RunName(ar.Int("run")) // used for naming logs, stats, etc
	if ar.Bool("epclog") {
		fnm := LogFileName("epc", netName, runName)
		logs.SetLogFile(etime.Train, etime.Epoch, fnm)
	}
	if ar.Bool("triallog") {
		fnm := LogFileName("trl", netName, runName)
		logs.SetLogFile(etime.Train, etime.Trial, fnm)
	}
	if ar.Bool("runlog") {
		fnm := LogFileName("run", netName, runName)
		logs.SetLogFile(etime.Train, etime.Run, fnm)
	}
	if ar.Bool("tstepclog") {
		fnm := LogFileName("tst_epc", netName, runName)
		logs.SetLogFile(etime.Test, etime.Epoch, fnm)
	}
	if ar.Bool("tsttriallog") {
		fnm := LogFileName("tst_trl", netName, runName)
		logs.SetLogFile(etime.Test, etime.Trial, fnm)
	}
}
