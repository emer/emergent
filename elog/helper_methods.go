package elog

import (
	"fmt"
	"github.com/emer/etable/etensor"
)

// ParamsName returns name of current set of parameters
func ParamsName(paramset string) string {
	if paramset == "" {
		return "Base"
	}
	return paramset
}

// RunName returns a name for this run that combines Tag and Params -- add this to
// any file names that are saved.
func RunName(tag string, paramName string) string { // TODO(library): library code as above
	if tag != "" {
		return tag + "_" + ParamsName(paramName)
	} else {
		return ParamsName(paramName)
	}
}

// RunEpochName returns a string with the run and epoch numbers with leading zeros, suitable
// for using in weights file names.  Uses 3, 5 digits for each.
func RunEpochName(run, epc int) string { // TODO(library): library, probably elog
	return fmt.Sprintf("%03d_%05d", run, epc)
}

// WeightsFileName returns default current weights file name
func WeightsFileName(netName, tag, paramName string, run, epc int) string { // TODO(library): library elog
	return netName + "_" + RunName(tag, paramName) + "_" + RunEpochName(run, epc) + ".wts.gz"
}

// LogFileName returns default log file name
func LogFileName(netName, lognm, tag, paramName string) string { // TODO(library): library elog
	return netName + "_" + RunName(tag, paramName) + "_" + lognm + ".tsv"
}

// ValsTsr gets value tensor of given name, creating if not yet made
func ValsTsr(tensorDictionary *map[string]*etensor.Float32, name string) *etensor.Float32 { // TODO(library): library code elog
	if *tensorDictionary == nil {
		*tensorDictionary = make(map[string]*etensor.Float32)
	}
	tsr, ok := (*tensorDictionary)[name]
	if !ok {
		tsr = &etensor.Float32{}
		(*tensorDictionary)[name] = tsr
	}
	return tsr
}
