#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJ_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

go build -o "${PROJ_ROOT}/out/xns-informer-gen" "${PROJ_ROOT}/cmd/xns-informer-gen"

function join_by { local IFS="$1"; shift; echo "$*"; }

# TODO: This is obviously a hack. This script needs to be updated to
# find these dynamically -- among other things.
k8s_group_versions=(
  k8s.io/api/admission/v1
  k8s.io/api/admission/v1beta1
  k8s.io/api/admissionregistration/v1
  k8s.io/api/admissionregistration/v1beta1
  k8s.io/api/apps/v1
  k8s.io/api/apps/v1beta1
  k8s.io/api/apps/v1beta2
  k8s.io/api/authentication/v1
  k8s.io/api/authentication/v1beta1
  k8s.io/api/authorization/v1
  k8s.io/api/authorization/v1beta1
  k8s.io/api/autoscaling/v1
  k8s.io/api/autoscaling/v2beta1
  k8s.io/api/autoscaling/v2beta2
  k8s.io/api/batch/v1
  k8s.io/api/batch/v1beta1
  k8s.io/api/batch/v2alpha1
  k8s.io/api/certificates/v1
  k8s.io/api/certificates/v1beta1
  k8s.io/api/coordination/v1
  k8s.io/api/coordination/v1beta1
  k8s.io/api/core/v1
  k8s.io/api/discovery/v1alpha1
  k8s.io/api/discovery/v1beta1
  k8s.io/api/events/v1
  k8s.io/api/events/v1beta1
  k8s.io/api/extensions/v1beta1
  k8s.io/api/flowcontrol/v1alpha1
  k8s.io/api/imagepolicy/v1alpha1
  k8s.io/api/networking/v1
  k8s.io/api/networking/v1beta1
  k8s.io/api/node/v1alpha1
  k8s.io/api/node/v1beta1
  k8s.io/api/policy/v1beta1
  k8s.io/api/rbac/v1
  k8s.io/api/rbac/v1alpha1
  k8s.io/api/rbac/v1beta1
  k8s.io/api/scheduling/v1
  k8s.io/api/scheduling/v1alpha1
  k8s.io/api/scheduling/v1beta1
  k8s.io/api/settings/v1alpha1
  k8s.io/api/storage/v1
  k8s.io/api/storage/v1alpha1
  k8s.io/api/storage/v1beta1
)

istio_group_versions=(
  istio.io/client-go/pkg/apis/networking/v1alpha3
  istio.io/client-go/pkg/apis/networking/v1beta1
  istio.io/client-go/pkg/apis/security/v1beta1
)

"${PROJ_ROOT}/out/xns-informer-gen" \
	-v 2 \
	-o "${PROJ_ROOT}/pkg/generated/kube" \
	-p 'github.com/maistra/xns-informer/pkg/generated/kube' \
	-i "$(join_by , ${k8s_group_versions[@]})"

"${PROJ_ROOT}/out/xns-informer-gen" \
  --listers-package "istio.io/client-go/pkg/listers" \
  --informers-package "istio.io/client-go/pkg/informers/externalversions" \
	-v 2 \
	-o "${PROJ_ROOT}/pkg/generated/istio" \
	-p 'github.com/maistra/xns-informer/pkg/generated/istio' \
	-i "$(join_by , ${istio_group_versions[@]})"
