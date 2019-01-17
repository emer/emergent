// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package leabra provides the basic reference leabra implementation, for rate-coded
activations and standard error-driven learning.  Other packages provide spiking
or deep leabra, PVLV, PBWM, etc.

The overall design seeks an "optimal" tradeoff between simplicity, transparency, ability to flexibly
recombine and extend elements, and avoiding having to rewrite a bunch of stuff.

The *Stru elements handle the core structural components of the network, and hold
emer.* interface pointers to elements such as emer.Layer, which provides a very minimal
interface for these elements.  Interfaces are automatically pointers, so think of these
as generic pointers to your specific Layers etc.

This design means the same *Stru infrastructure can be re-used across different variants
of the algorithm.  Because we're keeping this infrastructure minimal and algorithm-free
it should be much less confusing than dealing with the multiple levels of inheritance
in C++ emergent.  The actual algorithm-specific code is now fully self-contained,
and largely orthogonalized from the infrastructure.

One specific cost of this is the need to cast the emer.* interface pointers into
the specific types of interest, when accessing via the *Stru infrastructure.

*/
package leabra
