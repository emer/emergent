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

var KiT_TimeScales = kit.Enums.AddEnum(TimeScalesN, kit.NotBitFlag, nil)

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

	// Tick is one step in a sequence -- often it is useful to have Trial count
	// up throughout the entire Epoch but also include a Tick to count trials
	// within a Sequence
	Tick

	// Sequence is a sequential group of Trials (not always needed).
	Sequence

	// Block is a collection of Trials, Sequences or Events, often used in experiments
	// when conditions are varied across blocks.
	Block

	// Condition is a collection of Blocks that share the same set of parameters.
	// This is intermediate between Block and Run levels.
	Condition

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
