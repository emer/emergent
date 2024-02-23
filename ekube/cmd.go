// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"text/template"

	"cogentcore.org/core/xe"
)

// Build builds a Docker image for the emergent model in the current directory.
func Build(c *Config) error { //gti:add
	f, err := os.Create("Dockerfile")
	if err != nil {
		return err
	}
	defer f.Close()
	err = DockerfileTmpl.Execute(f, c)
	if err != nil {
		return err
	}
	return xe.Verbose().SetBuffer(false).Run("docker", "build", "-t", filepath.Base(c.Dir)+":latest", ".")
}

var DockerfileTmpl = template.Must(template.New("Dockerfile").Parse(
	`FROM golang:1.21-bookworm as builder
WORKDIR /build

COPY go.* ./
RUN go mod download
COPY . ./

RUN apt-get update && apt-get install -y libgl1-mesa-dev xorg-dev

RUN go build -o ./app ./{{.Dir}}

FROM scratch
WORKDIR /app
COPY --from=builder /build/app ./app
ENTRYPOINT ["./app"]
`))
