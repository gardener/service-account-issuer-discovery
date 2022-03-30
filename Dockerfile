# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.17.8 AS builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY app/ app/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o service-account-issuer-discovery main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/service-account-issuer-discovery .
# nonroot user https://github.com/GoogleContainerTools/distroless/blob/18b2d2c5ebfa58fe3e0e4ee3ffe0e2651ec0f7f6/base/base.bzl#L8
USER 65532:65532

ENTRYPOINT ["/service-account-issuer-discovery"]
