// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/kit"
)

// TimeScales are the different time scales associated with overall simulation running, and
// can be used to parameterize the updating and control flow of simulations at different scales.
// The definitions become increasingly subjective imprecise as the time scales increase.
// Environments can implement updating along different such time scales as appropriate.
// This list is designed to standardize terminology across simulations and
// establish a common conceptual framework for time -- it can easily be extended in specific
// simulations to add needed additional levels, although using one of the existing standard
// values is recommended wherever possible.
type TimeScales int32

//go:generate stringer -type=TimeScales

var KiT_TimeScales = kit.Enums.AddEnum(TimeScalesN, false, nil)

func (ev TimeScales) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *TimeScales) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The time scales
const (
	// Event is the smallest unit of naturalistic experience that coheres unto itself
	// (e.g., something that could be described in a sentence).
	// Typically this is on the time scale of a few seconds: e.g., reaching for
	// something, catching a ball.  In an experiment it could just be the onset
	// of a stimulus, or the generation of a response.
	Event TimeScales = iota

	// Trial is one unit of behavior in an experiment, and could potentially
	// encompass multiple Events (e.g., one event is fixation, next is stimulus,
	// last is response, all comprising one Trial).  It is also conventionally
	// used as a single Input / Output learning instance in a standard error-driven
	// learning paradigm.
	Trial

	// Sequence is a sequential group of Trials (not always needed).
	Sequence

	// Block is a collection of Trials, Sequences or Events, often used in experiments
	// when conditions are varied across blocks.
	Block

	// Epoch is used in two different contexts.  In machine learning, it represents a
	// collection of Trials, Sequences or Events that constitute a "representative sample"
	// of the environment.  In the simplest case, it is the entire collection of Trials
	// used for training.  In electrophysiology, it is a timing window used for organizing
	// the analysis of electrode data.
	Epoch

	// Run is a complete run of a model / subject, from training to testing, etc.
	// Often multiple runs are done in an Expt to obtain statistics over initial
	// random weights etc.
	Run

	// Expt is an entire experiment -- multiple Runs through a given protocol / set of
	// parameters.
	Expt

	// Scene is a sequence of events that constitutes the next larger-scale coherent unit
	// of naturalistic experience corresponding e.g., to a scene in a movie.
	// Typically consists of events that all take place in one location over
	// e.g., a minute or so. This could be a paragraph or a page or so in a book.
	Scene

	// Episode is a sequence of scenes that constitutes the next larger-scale unit
	// of naturalistic experience e.g., going to the grocery store or eating at a
	// restaurant, attending a wedding or other "event".
	// This could be a chapter in a book.
	Episode

	TimeScalesN
)

///////////////////////////////////////////////////////////////////////////
//  Ctr

// Ctr is a counter that counts increments at a given time scale.
// It keeps track of prior model queries and can compute when
// it has been incremented relative to that.
type Ctr struct {
	Scale     TimeScales `inactive:"+" desc:"the unit of time scale represented by this counter"`
	Cur       int        `desc:"current counter value"`
	Prv       int        `desc:"previous counter value, prior to last Incr() call"`
	PrevQuery int        `desc:"value of counter when it was last queried by the model"`
	Max       int        `desc:"where relevant, this is a fixed maximum counter value, above which the counter will reset back to 0 -- only used if > 0"`
}

// Init initializes counter -- Cur = 0 and PrevQuery = -1
func (ct *Ctr) Init() {
	ct.Prv = 0
	ct.Cur = 0
	ct.PrevQuery = 0
}

// Incr increments the counter by 1.  If Max > 0 then if Incr >= Max
// then the counter is reset to 0 and true is returned.  Otherwise false.
func (ct *Ctr) Incr() bool {
	ct.Prv = ct.Cur
	ct.Cur++
	if ct.Max > 0 && ct.Cur >= ct.Max {
		ct.Cur = 0
		return true
	}
	return false
}

// Query returns current counter value and a bool indicating whether the
// current value is different from whenit was previously queried, and
// updates the PrevQuery value to the current.
func (ct *Ctr) Query() (int, bool) {
	diff := ct.Cur != ct.PrevQuery
	ct.PrevQuery = ct.Cur
	return ct.Cur, diff
}

///////////////////////////////////////////////////////////////////////////
//  Utils

// SchemaFromScales returns an etable.Schema suitable for creating an
// etable.Table to record the given list of time scales.  Can then add
// to this schema anything else that might be needed, before using it
// to create a Table.
func SchemaFromScales(ts []TimeScales) etable.Schema {
	sc := make(etable.Schema, len(ts))
	for i, t := range ts {
		sc[i].Name = t.String()
		sc[i].Type = etensor.INT64
		sc[i].CellShape = nil
	}
	return sc
}
