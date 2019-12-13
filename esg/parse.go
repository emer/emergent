// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"bufio"
	"errors"
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

// ReadRules reads in a text file with rules, line-by-line simple parser
func (rls *Rules) ReadRules(r io.Reader) []error {
	var errs []error
	scan := bufio.NewScanner(r) // line at a time
	rstack := []*Rule{}
	lastwascmt := false
	lastcmt := ""
	for scan.Scan() {
		b := scan.Bytes()
		bs := string(b)
		sp := strings.Fields(bs)
		nsp := len(sp)
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
			lastwascmt = false // skip these
		case sp[0] == "}":
			lastwascmt = false
			sz := len(rstack)
			if sz == 0 {
				err := errors.New("esg.Rules parse error: mismatched end bracket } has no match")
				errs = append(errs, err)
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
				err := errors.New("esg.Rules parse error: start bracket { needs at least a rule name")
				errs = append(errs, err)
				continue
			}
			rnm := sp[nsp-2]
			cond := false
			if rnm == "?" {
				cond = true
				if nsp == 2 {
					err := errors.New("esg.Rules parse error: start cond bracket ? { needs at least a rule name")
					errs = append(errs, err)
					continue
				}
				rnm = sp[nsp-3]
			}
			sz := len(rstack)
			if sz > 0 {
				cr, ci := rls.ParseAddItem(rstack, &errs, sp)
				ci.SubRule = &Rule{Name: cr.Name + "SubRule", Desc: desc, IsConds: cond}
				rstack = append(rstack, ci.SubRule)
				ncond := nsp - 1
				if cond {
					ncond--
				}
				ci.Cond = rls.ParseConds(sp[:ncond], &errs)
			} else {
				nr := &Rule{Name: rnm, Desc: desc, IsConds: cond}
				rstack = append(rstack, nr)
				rls.Add(nr)
			}
		case sp[nsp-1] == "}":
			cr, ci := rls.ParseAddItem(rstack, &errs, sp)
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
			ci.Cond = rls.ParseConds(sp[:sbidx], &errs)
			it := &Item{}
			ci.SubRule.Items = append(ci.SubRule.Items, it)
			rls.ParseElems(ci.SubRule, it, sp[sbidx+1:nsp-1], &errs)
		case sp[0][0] == '=':
			rl := rls.ParseCurRule(rstack, &errs, sp)
			rls.ParseState(sp[0][1:], &rl.State, &errs)
		case sp[0][0] == '%':
			rl, it := rls.ParseAddItem(rstack, &errs, sp)
			if rl == nil {
				continue
			}
			pct, err := strconv.ParseFloat(sp[0][1:], 32)
			if err != nil {
				errs = append(errs, err)
			}
			it.Prob = float32(pct / 100)
			rl.HasProbs = true
			rls.ParseElems(rl, it, sp[1:], &errs)
		default:
			rl, it := rls.ParseAddItem(rstack, &errs, sp)
			if rl == nil {
				continue
			}
			rls.ParseElems(rl, it, sp, &errs)
		}
	}
	if len(errs) > 0 {
		fmt.Printf("\nParse errors:\n")
		for _, err := range errs {
			fmt.Println(err)
		}
	}
	return errs
}

func (rls *Rules) ParseCurRule(rstack []*Rule, errs *[]error, sp []string) *Rule {
	sz := len(rstack)
	if sz == 0 {
		err := fmt.Errorf("esg.Rules parse error: no active rule when defining items: %v", sp)
		*errs = append(*errs, err)
		return nil
	}
	return rstack[sz-1]
}

func (rls *Rules) ParseAddItem(rstack []*Rule, errs *[]error, sp []string) (*Rule, *Item) {
	rl := rls.ParseCurRule(rstack, errs, sp)
	if rl == nil {
		return nil, nil
	}
	it := &Item{}
	rl.Items = append(rl.Items, it)
	return rl, it
}

func (rls *Rules) ParseElems(rl *Rule, it *Item, els []string, errs *[]error) {
	for _, es := range els {
		switch {
		case es[0] == '=':
			rls.ParseState(es[1:], &it.State, errs)
		case es[0] == '\'':
			tok := es[1 : len(es)-1]
			it.Elems = append(it.Elems, Elem{El: TokenEl, Value: tok})
		default:
			it.Elems = append(it.Elems, Elem{El: RuleEl, Value: es})
		}
	}
}

func (rls *Rules) ParseState(ststr string, state *State, errs *[]error) {
	stsp := strings.Split(ststr, "=")
	if len(stsp) == 0 {
		err := fmt.Errorf("esg.Rules parse error: state expr: %v empty", ststr)
		*errs = append(*errs, err)
	} else {
		state.Name = stsp[0]
		if len(stsp) > 1 {
			state.Value = stsp[1]
		}
	}
}

func (rls *Rules) ParseConds(cds []string, errs *[]error) Conds {
	cs := Conds{}
	cur := &cs
	substack := []*Conds{cur}
	for _, c := range cds {
		for {
			csz := len(c)
			switch {
			case csz == 0:
				*errs = append(*errs, errors.New("esg.Rules parse error: no text left in cond expr"))
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
					*errs = append(*errs, errors.New("esg.Rules parse error: imbalanced parens in cond expr: "+strings.Join(cds, " ")))
				} else {
					*cur = append(*cur, &Cond{El: CRule, Rule: c[:csz-1]})
					cur = substack[ssz-2]
					substack = substack[:ssz-1]
				}
			case c == ")":
				ssz := len(substack)
				if ssz == 1 {
					*errs = append(*errs, errors.New("esg.Rules parse error: imbalanced parens in cond expr: "+strings.Join(cds, " ")))
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
