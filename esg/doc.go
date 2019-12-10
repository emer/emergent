// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package esg is the emergent stochastic generator, where tokens are generated
stochastically according to rules defining the contingencies and probabilities.
It can be used for generating sentences (sg as well).

There are two types of rules:
* unconditional random items
* conditional items.

Unconditional items are chosen at random, optionally with specified probabilities:

RuleName {
   %50 Rule2 Rule4
   %30 'token1' 'token2'
	...
}

where Items on separate lines within each rule are orthogonal options,
chosen at uniform random unless otherwise specified with a leading %pct.
%pct can add up to < 100 in which case *nothing* is an alternative output.

Multiple elements in an item (on the same line) are resolved and
emitted in order.
Terminal literals are 'quoted' -- otherwise refers to another rule:
error message will flag missing rules.

Conditional items are specified by the ? after the rule name:

RuleName ? {
   Rule2 || Rule3 {
		Item1
		Item2
		...
	}
	Rule5 && Rule6 {
		...
	}
	...
}

The expression before the opening bracket for each item is a standard logical expression
using || (or), && (and), and ! (not), along with parens,
where the elements are rules that could have been generated earlier in the pass --
they evaluate to true if so, and false if not.

If the whole expression evaluates to true, then it is among items chosen at random
(typically only one for conditionals but could be any number).

If just one item per rule it can be put all on one line.

Conventional modifiers, used for defining sub-rules:
A = Agent
	Ao = Co-Agent
V = Verb
P = Patient,
	Pi = Instrument
	Pc = Co-Patient
L = Location
R = adverb
*/
package esg
