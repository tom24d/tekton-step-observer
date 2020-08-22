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

REPO_ROOT_DIR=$(git rev-parse --show-toplevel)
# Vendored eventing test image.
readonly VENDOR_EVENTING_TEST_IMAGES="vendor/knative.dev/eventing/test/test_images/"
readonly PLUGIN_INSTALLATION_CONFIG=${REPO_ROOT_DIR}/config/
readonly VERSION_EVENTING="0.16.1"

source ${REPO_ROOT_DIR}/vendor/github.com/tektoncd/plumbing/scripts/e2e-tests.sh


# This vendors test image code from eventing, in order to use ko to resolve them into Docker images, the
# path has to be a GOPATH.
header "Publish test image"
sed -i 's@knative.dev/eventing/test/test_images@github.com/tom24d/step-observe-controller/vendor/knative.dev/eventing/test/test_images@g' "${VENDOR_EVENTING_TEST_IMAGES}"*/*.yaml
$(dirname $0)/upload-test-images.sh ${VENDOR_EVENTING_TEST_IMAGES} e2e || fail_test "Error uploading eventing test images"


header "Setting up environment"

echo "Installing tekton pipeline"
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
wait_until_pods_running tekton-pipelines || fail_test "tekton pipeline does not show up"

header "Install Eventing"
kubectl apply --filename https://github.com/knative/eventing/releases/download/v${VERSION_EVENTING}/eventing-crds.yaml
kubectl apply --filename https://github.com/knative/eventing/releases/download/v${VERSION_EVENTING}/eventing-core.yaml
kubectl apply --filename https://github.com/knative/eventing/releases/download/v${VERSION_EVENTING}/mt-channel-broker.yaml
kubectl apply --filename https://github.com/knative/eventing/releases/download/v${VERSION_EVENTING}/in-memory-channel.yaml
wait_until_pods_running knative-eventing || fail_test "Knative Eventing not up"


# set up plugin
echo "Installing step-observe-controller"
ko apply -f "${PLUGIN_INSTALLATION_CONFIG}"
wait_until_pods_running tekton-pipelines || fail_test "step-observe-controller does not show up"

# Run the integration tests
header "Running Go e2e tests"
go_test_e2e -timeout=15m ./test/... -channels=messaging.knative.dev/v1:InMemoryChannel || fail_test "go test failed"


(( failed )) && fail_test
success
