#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# For Linux
CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 |tr -d '\n')
export CA_BUNDLE
sed "s/caBundle: @@CA_BUNDLE@@/caBundle: ${CA_BUNDLE}/g" deploys/dev/webhook.yaml.tpl > deploys/dev/webhook.yaml

# For MacOS use the following
# CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 )
# export CA_BUNDLE
# sed "s/caBundle: @@CA_BUNDLE@@/caBundle: ${CA_BUNDLE}/g" deploys/dev/webhook.yaml.tpl > deploys/dev/webhook.yaml
