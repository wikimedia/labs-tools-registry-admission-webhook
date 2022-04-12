#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


help() {
    cat <<EOH
    Usage: $0 [OPTIONS] <ENVIRONMENT>

    Options:
      -h        Show this help.
      -c        Force refresh the certificates (only for new installations or
                if they are expired).
      -b        Also build the container (locally only).
      -v        Show verbose output.
      -i        If set, it will request user input before doing any changes.
      -f        If this is the first time deploying this to a running instance,
                this will take care of updating the existing resources to be
                able to deploy with helm

EOH
}


main () {
    local do_build="no"
    local refresh_certs="no"
    local interactive_flag=""
    local first_deploy="no"

    while getopts "hvcbif" option; do
        case "${option}" in
        h)
            help
            exit 0
            ;;
        b) do_build="yes";;
        c) refresh_certs="yes";;
        v) set -x;;
        i) interactive_flag="-i";;
        f) first_deploy="yes";;
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

    if [[ "$first_deploy" == "yes" ]]; then
        kubectl patch \
            ClusterRoleBinding registry-admission-psp \
            --patch-file ./deploy/first_deploy_patch.yaml
        kubectl patch \
            --namespace registry-admission \
            service registry-admission \
            --patch-file ./deploy/first_deploy_patch.yaml
        # prevent the misbehavior of the hook from deploying
        kubectl delete \
            ValidatingWebhookConfiguration registry-admission \
        || true
        # we need to recreate it as we are changing immutable fields (ex. selector)
        kubectl delete \
            --namespace registry-admission \
            deployment registry-admission \
        || true
    fi

    helmfile \
        $interactive_flag \
        --environment "$environment" \
        apply
}



main "$@"
