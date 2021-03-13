// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// OpenRules reads in a text file with rules, line-by-line simple parser
func (rls *Rules) OpenRules(fname string) []error {
	fp, err := os.Open(fname)
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return nil
	}
	return rls.ReadRules(fp)
}

// OpenRulesPy reads in a text file with rules, line-by-line simple parser
func (rls *Rules) OpenRulesPy(fname string) {
	rls.OpenRules(fname)
}

// AddParseErr adds given parser error, auto including line number
func (rls *Rules) AddParseErr(msg string) {
	err := fmt.Errorf("Line: %d \tesg Parse Error: %s", rls.ParseLn, msg)
	rls.ParseErrs = append(rls.ParseErrs, err)
}

// ReadRules reads in a text file with rules, line-by-line simple parser
func (rls *Rules) ReadRules(r io.Reader) []error {
	rls.Map = nil
	rls.Top = nil
	rls.ParseErrs = nil
	rls.ParseLn = 0
	scan := bufio.NewScanner(r) // line at a time
	rstack := []*Rule{}
	lastwascmt := false
	lastcmt := ""
	for scan.Scan() {
		rls.ParseLn++
		b := scan.Bytes()
		bs := string(b)
		sp := strings.Fields(bs)
		nsp := len(sp)
		if nsp > 2 && sp[0] != "//" { // get rid of trailing comments
			for i, s := range sp {
				if s == "//" {
					nsp = i
					sp = sp[:i]
					break
				}
			}
		}
		switch {
		case nsp == 0:
			lastwascmt = false
		case sp[0] == "//":
			ncmt := strings.Join(sp[1:], " ")
			if lastwascmt {
				lastcmt += "\n" + ncmt
			} else {
				lastcmt = ncmt
				lastwascmt = true
			}
		case len(sp[0]) > 2 && sp[0][:2] == "//":
			lastwascmt = false // repeated comment line skip these
		case sp[0] == "}":
			lastwascmt = false
			sz := len(rstack)
			if sz == 0 {
				rls.AddParseErr("mismatched end bracket } has no match")
				continue
			}
			rstack = rstack[:sz-1]
		case sp[nsp-1] == "{":
			desc := ""
			if lastwascmt {
				desc = lastcmt
				lastwascmt = false
			}
			if nsp == 1 {
				rls.AddParseErr("start bracket: '{' needs at least a rule name")
				continue
			}
			rnm := sp[0]
			var rptp float32
			prp := sp[nsp-2]
			if len(prp) > 2 && prp[0:2] == "=%" {
				pct, err := strconv.ParseFloat(prp[2:], 32)
				if err != nil {
					rls.AddParseErr(err.Error())
				} else {
					rptp = float32(pct / 100)
				}
			}
			typ := UniformItems
			switch prp {
			case "?":
				typ = CondItems
			case "|":
				typ = SequentialItems
			case "$":
				typ = PermutedItems
			}
			if typ != UniformItems {
				if nsp == 2 {
					rls.AddParseErr("start special bracket: '? {' needs at least a rule name")
					continue
				}
			}
			sz := len(rstack)
			if sz > 0 {
				cr, ci := rls.ParseAddItem(rstack, sp)
				ci.SubRule = &Rule{Name: cr.Name + "SubRule", Desc: desc, Type: typ, RepeatP: rptp}
				rstack = append(rstack, ci.SubRule)
				ncond := nsp - 1
				if typ == CondItems {
					ncond--
				}
				ci.Cond = rls.ParseConds(sp[:ncond])
			} else {
				nr := &Rule{Name: rnm, Desc: desc, Type: typ, RepeatP: rptp}
				rstack = append(rstack, nr)
				rls.Add(nr)
			}
		case sp[nsp-1] == "}":
			cr, ci := rls.ParseAddItem(rstack, sp)
			if cr == nil {
				continue
			}
			ci.SubRule = &Rule{Name: cr.Name + "SubRule"}
			sbidx := 0
			for si, s := range sp {
				if s == "{" {
					sbidx = si
				}
			}
			ci.Cond = rls.ParseConds(sp[:sbidx])
			it := &Item{}
			ci.SubRule.Items = append(ci.SubRule.Items, it)
			rls.ParseElems(ci.SubRule, it, sp[sbidx+1:nsp-1])
		case sp[0][0] == '=':
			rl := rls.ParseCurRule(rstack, sp)
			rls.ParseState(sp[0][1:], &rl.State)
		case sp[0][0] == '%':
			rl, it := rls.ParseAddItem(rstack, sp)
			if rl == nil {
				continue
			}
			pct, err := strconv.ParseFloat(sp[0][1:], 32)
			if err != nil {
				rls.AddParseErr(err.Error())
			}
			it.Prob = float32(pct / 100)
			if rl.Type == UniformItems {
				rl.Type = ProbItems
			}
			rls.ParseElems(rl, it, sp[1:])
		default:
			rl, it := rls.ParseAddItem(rstack, sp)
			if rl == nil {
				continue
			}
			rls.ParseElems(rl, it, sp)
		}
	}
	if len(rls.ParseErrs) > 0 {
		fmt.Printf("\nesg Parse errors for: %s\n", rls.Name)
		for _, err := range rls.ParseErrs {
			fmt.Println(err)
		}
	}
	return rls.ParseErrs
}

func (rls *Rules) ParseCurRule(rstack []*Rule, sp []string) *Rule {
	sz := len(rstack)
	if sz == 0 {
		rls.AddParseErr(fmt.Sprintf("no active rule when defining items: %v", sp))
		return nil
	}
	return rstack[sz-1]
}

func (rls *Rules) ParseAddItem(rstack []*Rule, sp []string) (*Rule, *Item) {
	rl := rls.ParseCurRule(rstack, sp)
	if rl == nil {
		return nil, nil
	}
	it := &Item{}
	rl.Items = append(rl.Items, it)
	return rl, it
}

func (rls *Rules) ParseElems(rl *Rule, it *Item, els []string) {
	for _, es := range els {
		switch {
		case es[0] == '=':
			rls.ParseState(es[1:], &it.State)
		case es[0] == '\'':
			if len(es) < 3 {
				rls.AddParseErr(fmt.Sprintf("empty token: %v in els: %v", es, els))
			} else {
				tok := es[1 : len(es)-1]
				it.Elems = append(it.Elems, Elem{El: TokenEl, Value: tok})
			}
		default:
			it.Elems = append(it.Elems, Elem{El: RuleEl, Value: es})
		}
	}
}

func (rls *Rules) ParseState(ststr string, state *State) {
	stsp := strings.Split(ststr, "=")
	if len(stsp) == 0 {
		rls.AddParseErr(fmt.Sprintf("state expr: %v empty", ststr))
	} else {
		if len(stsp) > 1 {
			state.Add(stsp[0], stsp[1])
		} else {
			state.Add(stsp[0], "")
		}
	}
}

func (rls *Rules) ParseConds(cds []string) Conds {
	cs := Conds{}
	cur := &cs
	substack := []*Conds{cur}
	for _, c := range cds {
		for {
			csz := len(c)
			switch {
			case csz == 0:
				rls.AddParseErr("no text left in cond expr")
			case c == "&&":
				*cur = append(*cur, &Cond{El: And})
			case c == "||":
				*cur = append(*cur, &Cond{El: Or})
			case c[0] == '!':
				*cur = append(*cur, &Cond{El: Not})
				c = c[1:]
				continue
			case c == "(":
				sub := &Cond{El: SubCond}
				*cur = append(*cur, sub)
				cur = &sub.Conds
				substack = append(substack, cur)
			case c[0] == '(':
				sub := &Cond{El: SubCond}
				*cur = append(*cur, sub)
				cur = &sub.Conds
				substack = append(substack, cur)
				c = c[1:]
				continue
			case c[csz-1] == ')':
				ssz := len(substack)
				if ssz == 1 {
					rls.AddParseErr("imbalanced parens in cond expr: " + strings.Join(cds, " "))
				} else {
					*cur = append(*cur, &Cond{El: CRule, Rule: c[:csz-1]})
					cur = substack[ssz-2]
					substack = substack[:ssz-1]
				}
			case c == ")":
				ssz := len(substack)
				if ssz == 1 {
					rls.AddParseErr("imbalanced parens in cond expr: " + strings.Join(cds, " "))
				} else {
					cur = substack[ssz-2]
					substack = substack[:ssz-1]
				}
			default:
				*cur = append(*cur, &Cond{El: CRule, Rule: c})
			}
			break
		}
	}
	return cs
}
