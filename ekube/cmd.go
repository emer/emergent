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

// Partially based on https://github.com/rickyjames35/vulkan_docker_test/blob/main/Dockerfile
var DockerfileTmpl = template.Must(template.New("Dockerfile").Parse(
	`FROM golang:1.21-bookworm as builder
WORKDIR /build
COPY . ./

WORKDIR /build/{{.Dir}}
RUN go build -tags offscreen -o ./app

FROM ubuntu:latest as runner

COPY --from=builder /build/{{.Dir}} /build

# Needed to share GPU
ENV NVIDIA_DRIVER_CAPABILITIES=all
ENV NVIDIA_VISIBLE_DEVICES=all

RUN apt-get update && \
	export DEBIAN_FRONTEND=noninteractive && \
	apt-get install -y pciutils vulkan-tools mesa-utils

WORKDIR /build
CMD ["./app", "-nogui"]
`))
