#!/usr/bin/env bash

APP_NAME=rayoperator
HEAD_NAME=rayhead
WORKER_NAME=rayworker

MANIFEST_DIR=./manifests

function update()
{
    # Setup Service Account
    kubectl $1 -f ${MANIFEST_DIR}/service_account.yaml
    kubectl $1 -f ${MANIFEST_DIR}/service.yaml

    # Setup RBAC
    kubectl $1 -f ${MANIFEST_DIR}/role.yaml
    kubectl $1 -f ${MANIFEST_DIR}/role_binding.yaml

    # Setup the CRD
    kubectl $1 -f ${MANIFEST_DIR}/crds/${APP_NAME}_v1alpha1_${HEAD_NAME}_crd.yaml
    kubectl $1 -f ${MANIFEST_DIR}/crds/${APP_NAME}_v1alpha1_${WORKER_NAME}_crd.yaml

    # Deploy the app-operator
    kubectl $1 -f ${MANIFEST_DIR}/operator.yaml

    # Create an AppService CR
    # The default controller will watch for AppService objects and create a pod for each CR
    kubectl $1 -f ${MANIFEST_DIR}/crds/${APP_NAME}_v1alpha1_${HEAD_NAME}_cr.yaml
    kubectl $1 -f ${MANIFEST_DIR}/crds/${APP_NAME}_v1alpha1_${WORKER_NAME}_cr.yaml

    # Verify that a pod is created
    kubectl get pod
}

function stop()
{
    update "delete"
}

function deploy()
{
    update "create"
}
