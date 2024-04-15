// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command ekube provides easy building of Docker images for emergent models
// and the deployment of those images to Kubernetes clusters.
package main

import "cogentcore.org/core/cli"

//go:generate core generate

func main() {
	opts := cli.DefaultOptions("ekube", "ekube provides easy building of Docker images for emergent models and the deployment of those images to Kubernetes clusters.")
	cli.Run(opts, &Config{}, Build)
}
