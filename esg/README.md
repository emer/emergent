[GoDoc](https://godoc.org/github.com/emer/emergent/esg)

Package esg is the emergent stochastic generator, where tokens are generated stochastically according to rules defining the contingencies and probabilities.  It can be used for generating sentences (sg as well).

# Rules

There are 5 types of rules, based on how the items within the rule are selected:
* Uniform random and random with specified probabilities.
* Conditional items that depend on a logical expression -- use the `?` before `{` to mark.
* Sequential and permuted order items that iterate through the list -- use `|` or `$` respectively.

Unconditional items are chosen at random, optionally with specified probabilities:

```
RuleName {
    %50 Rule2 Rule4
    %30 'token1' 'token2'
    ...
}
```

where Items on separate lines within each rule are orthogonal options, chosen at uniform random unless otherwise specified with a leading %pct. %pct can add up to < 100 in which case *nothing* is an alternative output.

Multiple elements in an item (on the same line) are resolved and emitted in order. Terminal literals are 'quoted' -- otherwise refers to another rule: error message will flag missing rules during Validate().

Conditional items are specified by the ? after the rule name:

```
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
```

The expression before the opening bracket for each item is a standard logical expression using || (or), && (and), and ! (not), along with parens, where the elements are rules or output tokens (which must be enclosed in ' ' single quotes) that could have been generated earlier in the pass -- they evaluate to true if so, and false if not.

If the whole expression evaluates to true, then it is among items chosen at random (typically only one for conditionals but could be any number).

If just one item per rule it can be put all on one line.

# Repeating choices over time:

Any rule can have an optional `=%p` expression just before the `{`, which indicates the probability of repeating the same item as last time:

```
RuleName ? =%70 {
...
```

This gives the rule a 70% chance of repeating the same item, regardless of how it was chosen before.  Note that this probably doesn't make a lot of sense for conditional rules as the choice last time may not satisfy the conditions this time.

# States

Each Rule or Item can have an optional State expression associated with it, which will update a `States` map in the overall Rules if that Rule or Item fires.  The `States` map is a simple `map[string]string` and is reset for every Gen pass.  It is useful for annotating the underlying state of the world implied by the generated sentence, using various role-filler values, such as those given by the modifiers below.  Although the Fired map can be used to recover this information, the States map and expressions can be designed to make a clean, direct state map with no ambiguity.

In the rules text file, an `=` prefix indicates a state-setting expression -- it can either be a full expression or get the value automatically:
* `=Name=Value` -- directly sets state Name to Value
* `=Name`  -- sets state Name to value of Item or Rule that it is associated with.  Only non-conditional Items can be used to set the value, which is the first element in the item expression -- conditionals with sub-rules must set the value explicitly.

Expressions at the start of a rule (or sub-rule), on a separate line, are associated with the rule and activate when that rule is fired (and the implicit value is the name of the rule).  Expressions at the end of an Item line are associated with the Item.  Put single-line sub-rule state expressions at end just before `}`.  Any number of state expressions can be added.

# Std Modifiers

Conventional modifiers, used for defining sub-rules:
* A = Agent
    + Ao = CoAgent
* V = Verb
* P = Patient,
    + Pi = Instrument
    + Pc = CoPatient
* L = Location
* R = Adverb

See [testdata/testrules.txt](https://github.com/emer/emergent/blob/master/esg/testdata/testrules.txt) and [sg CCN sim](https://github.com/CompCogNeuro/sims/blob/master/ch9/sg) for example usage.

