#!/usr/bin/env bash

set -x

# This script should be run inside the /hack

ROOT_DIR=../../
DEST_DIR=../../

APP_NAME=ray-operator
APP_DOMAIN=ray-operator.io
WORKER_CRD=RayWorker
HEAD_CRD=RayHead

IMAGE_REPO=rayop  # default to docker hub
IMAGE_NAME=${APP_NAME}

function create_operator() {
    echo "creating operator.."
    #rm -r ${ROOT_DIR} || true
    mkdir -p ${ROOT_DIR}

    pushd .
    cd ${ROOT_DIR}

    echo "using operator-sdk version:"
    operator-sdk version
    export GO111MODULE=on

    operator-sdk new ${APP_NAME} --dep-manager=dep
    cd ${APP_NAME}

    # add crd
    operator-sdk add api --api-version=${APP_DOMAIN}/v1alpha1 --kind=${WORKER_CRD}
    operator-sdk add api --api-version=${APP_DOMAIN}/v1alpha1 --kind=${HEAD_CRD}

    # add controller
    operator-sdk add controller --api-version=${APP_DOMAIN}/v1alpha1 --kind=${WORKER_CRD}
    operator-sdk add controller --api-version=${APP_DOMAIN}/v1alpha1 --kind=${HEAD_CRD}
    popd
}

function create_image() {
    echo "creating operator imagej.."
    pushd .
    cd ${ROOT_DIR}/${APP_NAME}

    # build and push operator image
    operator-sdk build ${IMAGE_REPO}/${IMAGE_NAME}

    docker login -u ${IMAGE_REPO}
    docker push ${IMAGE_REPO}/${IMAGE_NAME}

    # Update the operator manifest to use the built image name
    if [[ ${OSTYPE} = *darwin* ]]; then
        sed -i "" 's|REPLACE_IMAGE|'"${IMAGE_REPO}/${IMAGE_NAME}"'|g' deploy/operator.yaml
    else
        sed -i 's|REPLACE_IMAGE|'"${IMAGE_REPO}/${IMAGE_NAME}"'|g' deploy/operator.yaml
    fi
    popd
}

function apply() {
    rm -r ${DEST_DIR} || true
    mv ${ROOT_DIR}/${APP_NAME} ${DEST_DIR}
}

create_operator
create_image
# apply

