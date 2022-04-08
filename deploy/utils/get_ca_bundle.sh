#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


main() {
    local kubectl
    local trim
    if type minikube > /dev/null; then
        kubectl="minikube kubectl"
    else
        kubectl="kubectl"
    fi

    if [[ "$OSTYPE" == 'darwin'* ]]; then
        trim='cat'
    else
        trim='tr -d \n'
    fi
    # shellcheck disable=SC2086
    $kubectl -- \
        get configmap \
        -n kube-system \
        extension-apiserver-authentication \
        -o=jsonpath='{.data.client-ca-file}' \
    | base64 - \
    | $trim
}


main "$@"
