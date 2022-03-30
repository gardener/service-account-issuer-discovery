# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

HELM              ?= helm
HELM_CHART_DIR    ?= charts/service-account-issuer-discovery

.PHONY: helm-lint
helm-lint:
	@$(HELM) lint $(HELM_CHART_DIR)
