#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJ_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

go build -o "${PROJ_ROOT}/out/xns-informer-gen" "${PROJ_ROOT}/cmd/xns-informer-gen"

"${PROJ_ROOT}/out/xns-informer-gen" \
	-v 2 \
	-o "${PROJ_ROOT}/pkg/informers/kube" \
	-p 'github.com/maistra/xns-informer/pkg/informers/kube' \
	-i 'k8s.io/api/core/v1'
