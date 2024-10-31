# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

HELM              ?= helm
HELM_CHART_DIR    ?= charts/service-account-issuer-discovery

.PHONY: helm-lint
helm-lint:
	@$(HELM) lint $(HELM_CHART_DIR)

.PHONY: check
check:
	@go vet ./...
	@go fmt ./...

.PHONY: test
test:
	@go test -cover ./...

.PHONY: verify
verify: check test sast

.PHONY: sast
sast:
	@./hack/sast.sh

.PHONY: sast-report
sast-report:
	@./hack/sast.sh --gosec-report true
