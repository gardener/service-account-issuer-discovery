# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

PACKAGE_PATH="${1:-github.com/gardener/service-account-issuer-discovery/internal}"
VERSION_PATH="${2:-$(dirname $0)/../VERSION}"
VERSION_VERSIONFILE="$(cat "$VERSION_PATH")"
VERSION="${VERSION_VERSIONFILE}-$(git rev-parse HEAD)"

if [ "$(git status --porcelain 2>/dev/null | wc -l)" -gt 0 ]
then
	VERSION=${VERSION}-dirty
fi

echo "-X $PACKAGE_PATH/version.version=$VERSION"
