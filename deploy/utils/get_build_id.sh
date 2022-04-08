#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


main() {
    local git_hash
    local suffix="${1:+-$1}"
    git_hash="$(git rev-parse HEAD)"
    echo "${git_hash}-$(date +%Y%m%d_%H%M%S)${suffix}"
}


main "$@"
