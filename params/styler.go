// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import "strings"

// The params.Styler interface exposes TypeName, Class, and Name methods
// that allow the params.Sel CSS-style selection specifier to determine
// whether a given parameter applies.
// Adding Set versions of Name and Class methods is a good idea but not
// needed for this interface, so they are not included here.
type Styler interface {
	// StyleType returns the name of this type for CSS-style matching.
	// This is used for CSS Sel selector with no prefix.
	// This type is used *in addition* to the actual Go type name
	// of the object, and is a kind of type-category (e.g., Layer
	// or Path in emergent network objects).
	StyleType() string

	// StyleClass returns the space-separated list of class selectors (tags).
	// Parameters with a . prefix target class tags.
	// Do NOT include the . in the Class tags on Styler objects;
	// The . is only used in the Sel selector on the params.Sel.
	StyleClass() string

	// StyleName returns the name of this object.
	// Parameters with a # prefix target object names, which are typically
	// unique.  Note, do not include the # prefix in the actual object name,
	// only in the Sel selector on params.Sel.
	StyleName() string
}

// The params.StylerObject interface extends Styler to include an arbitary
// function to access the underlying object type.
type StylerObject interface {
	Styler

	// StyleObject returns the object that will have its field values set by
	// the params specifications.
	StyleObject() any
}

// AddClass adds given class(es) to current class string,
// ensuring it is not a duplicate of existing, and properly
// adding spaces
func AddClass(cur string, class ...string) string {
	cls := strings.Join(class, " ")
	if ClassMatch(cur, cls) {
		return cur
	}
	cur = strings.TrimSpace(cur)
	if len(cur) == 0 {
		return cls
	}
	return cur + " " + cls
}
