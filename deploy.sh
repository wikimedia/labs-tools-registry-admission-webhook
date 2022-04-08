#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


help() {
    cat <<EOH
    Usage: $0 [OPTIONS] <ENVIRONMENT>

    Options:
      -h        Show this help.
      -c        Force refresh the certificates (only for new installations or if they are expired).
      -b        Also build the container (locally only).
      -v        Show verbose output.

EOH
}


main () {
    local do_build="no"
    local refresh_certs="no"

    while getopts "hvcb" option; do
        case "${option}" in
        h)
            help
            exit 0
            ;;
        b) do_build="yes";;
        c) refresh_certs="yes";;
        v) set -x;;
        *)
            echo "Wrong option $option"
            help
            exit 1
            ;;
        esac
    done
    shift $((OPTIND-1))

    local environment="${1?No environment passed}"
    if [[ ! -f "deploy/values-$environment.yaml.gotmpl"  ]]; then
        echo "Unknown environment $environment, use one of:"
        find deploy \
            -iname 'values-*yaml.gotmpl' \
            | grep -v 'common' \
            | sed -e 's|deploy/values-\(.*\).yaml.gotmpl|\1|'
        exit 1
    fi

    if [[ "$do_build" == "yes" ]]; then
        if [[ "$environment" != "dev" ]]; then
            echo "You probably don't want this, as it will build the image locally, hit enter if you are sure. (hit Ctrl+C to cancel)"
            read -r
        fi
        # shellcheck disable=SC2046
        eval $(minikube docker-env)
        docker build . -t docker-registry.tools.wmflabs.org/registry-admission:latest
    fi

    if ! kubectl --namespace registry-admission get secret registry-admission-certs > /dev/null 2>&1; then
        refresh_certs="yes"
    fi
    if [[ "$refresh_certs" == "yes" ]]; then
        ./deploy/utils/get-cert.sh
    fi

    helmfile \
        --environment "$environment" \
        apply
}



main "$@"
