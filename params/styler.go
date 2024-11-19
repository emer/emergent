// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import "strings"

// Styler must be implemented by any object that parameters are
// applied to, to provide the .Class and #Name selector functionality.
type Styler interface {
	// StyleClass returns the space-separated list of class selectors (tags).
	// Parameters with a . prefix target class tags.
	// Do NOT include the . in the Class tags on Styler objects;
	// The . is only used in the Sel selector on the [Sel].
	StyleClass() string

	// StyleName returns the name of this object.
	// Parameters with a # prefix target object names, which are typically
	// unique. Do NOT include the # prefix in the actual object name,
	// which is only present in the Sel selector on [Sel].
	StyleName() string
}

// AddClass is a helper function that adds given class(es) to current
// class string, ensuring it is not a duplicate of existing, and properly
// adding spaces.
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
