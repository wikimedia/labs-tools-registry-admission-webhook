#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


help() {
    cat <<EOH
    Usage: $0 [OPTIONS] <ENVIRONMENT>

    Options:
      -h        Show this help.
      -c        Refresh also the certificates (only for new installations or if they are expired).
      -b        Also build the container (locally only).
      -v        Show verbose output.

EOH
}


deploy_generic() {
    local environment="${1?No environment passed}"
    kubectl apply -k "deploys/$environment"
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
    if [[ ! -d "deploys/$environment"  ]]; then
        echo "Unknown environment $environment, use one of:"
        ls deploys/
        exit 1
    fi

    if [[ "$do_build" == "yes" ]]; then
        if [[ "$environment" != "dev" ]]; then
            echo "You probably don't want this, as it will build the image locally, hit enter if you are sure. (hit Ctrl+C to cancel)"
            read -r
        fi
        # shellcheck disable=SC2046
        eval $(minikube docker-env)
        docker build . -t registry-admission:latest
    fi
    if [[ "$environment" == "dev" ]]; then
        ./ca-bundle.sh
    fi
    if [[ "$refresh_certs" == "yes" ]]; then
        ./get-cert.sh
    fi

    deploy_generic "$environment"
}



main "$@"
