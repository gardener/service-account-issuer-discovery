#!/usr/bin/env bash
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -e

PACKAGE_PATH="${1:-github.com/gardener/service-account-issuer-discovery/internal}"
VERSION_PATH="${2:-$(dirname $0)/../VERSION}"
VERSION_VERSIONFILE="$(cat "$VERSION_PATH")"
VERSION="${EFFECTIVE_VERSION:-$VERSION_VERSIONFILE}"

if [ "$(git status --porcelain 2>/dev/null | wc -l)" -gt 0 ]
then
	VERSION=${VERSION}-dirty
fi

echo "-X $PACKAGE_PATH/version.Version=$VERSION"
