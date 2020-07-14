#!/usr/bin/env bash

# Copyright 2019 The Tekton Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT_DIR=$(git rev-parse --show-toplevel)
readonly PLUGIN_INSTALLATION_CONFIG=${REPO_ROOT_DIR}/config/
readonly VERSION_TEKTON="0.14.1"

source ${REPO_ROOT_DIR}/vendor/github.com/tektoncd/plumbing/scripts/e2e-tests.sh

# Script entry point.
#initialize $@

header "Setting up environment"

echo "Installing tekton pipeline"
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v${VERSION_TEKTON}/release.yaml
wait_until_pods_running tekton-pipelines || fail_test "tekton pipeline does not show up"


# set up plugin
echo "Installing step-observe-controller"
ko apply -f "${PLUGIN_INSTALLATION_CONFIG}"
wait_until_pods_running tekton-pipelines || fail_test "step-observe-controller does not show up"

failed=0

# Run the integration tests
#header "Running Go e2e tests"
#go_test_e2e -timeout=20m ./test/... || failed=1

# Run these _after_ the integration tests b/c they don't quite work all the way
# and they cause a lot of noise in the logs, making it harder to debug integration
# test failures.
#go_test_e2e -tags=examples -timeout=20m ./test/ || failed=1

(( failed )) && fail_test
success
