#!/bin/bash

KUBECTL_VERSION="v1.11.1"
KUBECTL_URL="https://storage.googleapis.com"
KUBECTL_URL_PATH="kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
KUBECTL_TARGET_DIR="${KUBECTL_TARGET_DIR:-/home/travis/bin}"
KUBECTL_BIN="${KUBECTL_TARGET_DIR}/kubectl"

function die () {
    echo "[ERROR] ${*}" 1>&2
    exit 1
}

if ! curl --location --output ${KUBECTL_BIN} ${KUBECTL_URL}/${KUBECTL_URL_PATH} ; then
    die "On downloadig kubectl"
fi

chmod +x ${KUBECTL_BIN} || die "On changing kubectl mode"

if ! go get sigs.k8s.io/kind ; then
    die "On installing KinD"
fi

if ! kind create cluster ; then
    die "On creating a kind cluster"
fi

if [ ! -d "${HOME}/.kube" ] ; then
    mkdir ${HOME}/.kube || die "On creating '${HOME}/.kube' dir"
fi
