#!/bin/bash

set -eo pipefail

die () {
    echo >&2 "$@"
    show_help
    exit 1
}

function header {
    echo -e "\e[92m\e[4m\e[1m${1}\e[0m"
}

dryRun=false
skipInDryRun() {
  if $dryRun; then echo "# $*";  fi
  if ! $dryRun; then "$@";  fi
}

function isInArray {
    local value=$1
    local deps=$2
    for dep in "${deps[@]}"; do
        found=$(echo "$dep" | grep "$value" || true)
        if [ -n "$found" ]; then 
            return 0
        fi
    done
    return 1
}

show_help() {
  local usage
  usage="
$(basename "$0")

Updates dependecies of this project with versions from istio/istio. 
Takes into account replace and exclude directives from the upstream go.mod.
Modifies go.{mod,sum} as a result.

Usage:
  ./$(basename "$0") [flags]
   
Options:
  --version of istio/istio upstream project from which go.mod is used.
            Can be a branch, tag or any arbitrary sha.
 
  -d, --dry-run
    Does not execute actual commands but prints them instead.

  -h, --help       
    Help message.

Example:
  ./$(basename "$0") --version release-1.16
"

  echo "$usage"
}

while test $# -gt 0; do
  case "$1" in
    -h|--help)
            show_help
            exit 0
            ;;
    -d|--dry-run)
            dryRun=true
            shift
            ;;
    --version)
            if [[ $1 == "--"* ]]; then
                version="${2/--/}"
                shift
            fi
            shift
            ;;
    *)
            die "$(basename "$0"): unknown flag $(echo "$1" | cut -d'=' -f 1)"
            exit 1
            ;;
  esac
done

if [ -z "$version" ]; then
    die "Missing version"
fi

if ! command -v curl &>/dev/null; then
  echo "curl is required"
  exit 1
fi


deeptreeCmd="go run -mod=readonly github.com/vc60er/deptree@3cc6257204e26c85723f8e87bdfd9e63f24f2910"

istioDeps=$(curl -sfL https://raw.githubusercontent.com/istio/istio/"${version}"/go.mod)

mapfile -t deps < <(go mod graph | bash -c "${deeptreeCmd} -d 1" | cut -d' ' -f 2 | tr -s '\n' | sort | grep -v "tree:")
mapfile -t replaceDeps < <(echo "${istioDeps}" | grep -Po 'replace \K.*')
mapfile -t excludeDeps < <(echo "${istioDeps}" | grep -Po 'exclude \K.*' | sed -e "s/ /@/g")

header "Updating dependencies from istio@${version}"
for dep in "${deps[@]}"; do
    name="${dep%@*}" # dependency in graph is stored as name@version - take name from the variable
    istioDep=$(echo "${istioDeps}" | grep -v "replace" | grep -v "exclude" | grep "${name} " || true)
    if [ -n "$istioDep" ]; then    
        newVersion=${istioDep#*\ } # take new version from istio go.mod (it's a text file)
        skipInDryRun go mod edit -require="${name}@${newVersion%"// indirect"}" # strip off "// indirect" comment if any
    fi
done

header "Adding explicit replaces from istio@${version} go.mod"
for dep in "${replaceDeps[@]}"; do
    name=${dep%%\ *}
    newVersion=${dep##*\ }
    if isInArray "${name}" "${deps[*]}"; then
        skipInDryRun go mod edit -replace="${name}=${name}@${newVersion}"
    fi    
done

header "Adding explicit excludes from istio@${version} go.mod"
for dep in "${excludeDeps[@]}"; do
    skipInDryRun go mod edit -exclude="${dep}"
done

header "Updating go.sum"
skipInDryRun go mod tidy -compat=1.19
