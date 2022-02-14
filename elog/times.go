// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import "github.com/goki/ki/kit"

// Times the enum
type Times int32

//go:generate stringer -type=Times

var KiT_Times = kit.Enums.AddEnum(TimesN, kit.NotBitFlag, nil)

func (ev Times) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Times) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// A list of predefined time scales at which logging can occur
const (
	// NoTime represents a non-initialized value
	NoTime Times = iota

	// AllTimes indicates that the log should occur over all times present in other items.
	AllTimes

	// Cycle is the finest time scale -- typically 1 msec -- a single activation update.
	Cycle

	// FastSpike is typically 10 cycles = 10 msec (100hz) = the fastest spiking time
	// generally observed in the brain.  This can be useful for visualizing updates
	// at a granularity in between Cycle and GammaCycle.
	FastSpike

	// GammaCycle is typically 25 cycles = 25 msec (40hz)
	GammaCycle

	// Phase is either Minus or Plus phase, where plus phase is bursting / outcome
	// that drives positive learning relative to prediction in minus phase.
	Phase

	// BetaCycle is typically 50 cycles = 50 msec (20 hz) = one beta-frequency cycle.
	// Gating in the basal ganglia and associated updating in prefrontal cortex
	// occurs at this frequency.
	BetaCycle

	// AlphaCycle is typically 100 cycles = 100 msec (10 hz) = one alpha-frequency cycle.
	AlphaCycle

	// ThetaCycle is typically 200 cycles = 200 msec (5 hz) = two alpha-frequency cycles.
	// This is the modal duration of a saccade, the update frequency of medial temporal lobe
	// episodic memory, and the minimal predictive learning cycle (perceive an Alpha 1, predict on 2).
	ThetaCycle

	// Event is the smallest unit of naturalistic experience that coheres unto itself
	// (e.g., something that could be described in a sentence).
	// Typically this is on the time scale of a few seconds: e.g., reaching for
	// something, catching a ball.
	Event

	// Trial is one unit of behavior in an experiment -- it is typically environmentally
	// defined instead of endogenously defined in terms of basic brain rhythms.
	// In the minimal case it could be one ThetaCycle, but could be multiple, and
	// could encompass multiple Events (e.g., one event is fixation, next is stimulus,
	// last is response)
	Trial

	// Tick is one step in a sequence -- often it is useful to have Trial count
	// up throughout the entire Epoch but also include a Tick to count trials
	// within a Sequence
	Tick

	// Sequence is a sequential group of Trials (not always needed).
	Sequence

	// Condition is a collection of Blocks that share the same set of parameters.
	// This is intermediate between Block and Run levels.
	Condition

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

	TimesN
)
