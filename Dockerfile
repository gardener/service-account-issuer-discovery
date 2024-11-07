# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.23.3 AS builder

WORKDIR /workspace
COPY . .
RUN go mod download

# Build
ARG TARGETARCH
WORKDIR /workspace/cmd/service-account-issuer-discovery
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GO111MODULE=on go build -a -ldflags="$(/workspace/hack/get-build.sh)" -o /workspace/service-account-issuer-discovery

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/service-account-issuer-discovery /service-account-issuer-discovery
# nonroot user https://github.com/GoogleContainerTools/distroless/blob/18b2d2c5ebfa58fe3e0e4ee3ffe0e2651ec0f7f6/base/base.bzl#L8
USER 65532:65532

ENTRYPOINT ["/service-account-issuer-discovery"]
