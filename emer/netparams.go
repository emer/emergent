// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"log"
	"strings"

	"cogentcore.org/core/base/mpi"
	"github.com/emer/emergent/v2/netparams"
	"github.com/emer/emergent/v2/params"
)

// NetParams handles standard parameters for a Network only
// (use econfig and a Config struct for other configuration params)
// Assumes a Set named "Base" has the base-level parameters, which are
// always applied first, followed optionally by additional Set(s)
// that can have different parameters to try.
type NetParams struct {

	// full collection of param sets to use
	Params netparams.Sets `display:"no-inline"`

	// optional additional sheets of parameters to apply after Base -- can use multiple names separated by spaces (don't put spaces in Sheet names!)
	ExtraSheets string

	// optional additional tag to add to file names, logs to identify params / run config
	Tag string

	// the network to apply parameters to
	Network Network `display:"-"`

	// list of hyper parameters compiled from the network parameters, using the layers and pathways from the network, so that the same styling logic as for regular parameters can be used
	NetHypers params.Flex `display:"-"`

	// print out messages for each parameter that is set
	SetMsg bool
}

// Config configures the ExtraSheets, Tag, and Network fields
func (pr *NetParams) Config(pars netparams.Sets, extraSheets, tag string, net Network) {
	pr.Params = pars
	report := ""
	if extraSheets != "" {
		pr.ExtraSheets = extraSheets
		report += " ExtraSheets: " + extraSheets
	}
	if tag != "" {
		pr.Tag = tag
		report += " Tag: " + tag
	}
	pr.Network = net
	if report != "" {
		mpi.Printf("NetParams Set: %s", report)
	}
}

// Name returns name of current set of parameters, including Tag.
// if ExtraSheets is empty then it returns "Base", otherwise returns ExtraSheets
func (pr *NetParams) Name() string {
	rn := ""
	if pr.Tag != "" {
		rn += pr.Tag + "_"
	}
	if pr.ExtraSheets == "" {
		rn += "Base"
	} else {
		rn += pr.ExtraSheets
	}
	return rn
}

// RunName returns standard name simulation run based on params Name()
// and starting run number if > 0 (large models are often run separately)
func (pr *NetParams) RunName(startRun int) string {
	rn := pr.Name()
	if startRun > 0 {
		rn += fmt.Sprintf("_%03d", startRun)
	}
	return rn
}

// Validate checks that the Network has been set
func (pr *NetParams) Validate() error {
	if pr.Network == nil {
		err := fmt.Errorf("emer.NetParams: Network is not set -- params will not be applied!")
		log.Println(err)
		return err
	}
	return nil
}

// SetAll sets all parameters, using "Base" Set then any ExtraSheets,
// Does a Validate call first.
func (pr *NetParams) SetAll() error {
	err := pr.Validate()
	if err != nil {
		return err
	}
	if hist, ok := pr.Network.(params.History); ok {
		hist.ParamsHistoryReset()
	}
	err = pr.SetAllSheet("Base")
	if pr.ExtraSheets != "" && pr.ExtraSheets != "Base" {
		sps := strings.Fields(pr.ExtraSheets)
		for _, ps := range sps {
			err = pr.SetAllSheet(ps)
		}
	}
	return err
}

// SetAllSheet sets parameters for given Sheet name to the Network
func (pr *NetParams) SetAllSheet(sheetName string) error {
	err := pr.Validate()
	if err != nil {
		return err
	}
	psheet, err := pr.Params.SheetByNameTry(sheetName)
	if err != nil {
		return err
	}
	psheet.SelMatchReset(sheetName)
	pr.SetNetworkSheet(pr.Network, psheet, sheetName)
	err = psheet.SelNoMatchWarn(sheetName, "Network")
	return err
}

// SetNetworkMap applies params from given map of values
// The map keys are Selector:Path and the value is the value to apply, as a string.
func (pr *NetParams) SetNetworkMap(net Network, vals map[string]any) error {
	sh, err := params.MapToSheet(vals)
	if err != nil {
		log.Println(err)
		return err
	}
	pr.SetNetworkSheet(net, sh, "ApplyMap")
	return nil
}

// SetNetworkSheet applies params from given sheet
func (pr *NetParams) SetNetworkSheet(net Network, sh *params.Sheet, setName string) {
	net.ApplyParams(sh, pr.SetMsg)
	hypers := NetworkHyperParams(net, sh)
	if setName == "Base" {
		pr.NetHypers = hypers
	} else {
		pr.NetHypers.CopyFrom(hypers)
	}
}
