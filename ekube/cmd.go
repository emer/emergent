// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"text/template"

	"cogentcore.org/core/strcase"
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
	return xe.Verbose().SetBuffer(false).Run("docker", "build", "-t", strcase.ToKebab(c.Dir)+":latest", ".")
}

var DockerfileTmpl = template.Must(template.New("Dockerfile").Parse(
	`FROM golang:1.21-bookworm as builder
WORKDIR /app

COPY go.* ./
RUN go mod download
COPY . ./

RUN go build -tags offscreen -o app ./{{.Dir}}

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
	ca-certificates && \
	rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app /app/app

CMD ["/app/app"]
`))
