// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package emer provides minimal interfaces for the basic structural elements of neural networks
including:
* emer.Network, emer.Layer, emer.Unit, emer.Prjn (projection that interconnects layers)

These interfaces are intended to be just sufficient to support visualization and generic
analysis kinds of functions, but explicitly avoid exposing ANY of the algorithmic aspects,
so that those can be purely encoded in the implementation structs.

At this point, given the extra complexity it would require, these interfaces do not support
the ability to build or modify networks.

*/
package emer
