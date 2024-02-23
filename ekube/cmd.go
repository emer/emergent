// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "text/template"

// Build builds a Docker image for the emergent model in the current directory.
func Build(c *Config) error { //gti:add
	return nil
}

var DockerfileTmpl = template.Must(template.New("Dockerfile").Parse(``))
