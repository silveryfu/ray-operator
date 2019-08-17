## Overview

The Kubernetes Operator for Ray aims to make running [Ray](https://github.com/ray-project/ray) on Kubernetes easy and efficient. The goal is to support the following list of features:

* Declarative specification of Ray workers and applications;
* Automatic restarts of Ray workers and retries of Ray tasks;
* Automatic clean-up/unprovisioning;
* Autoscaling Ray cluster

## Run ray-operator

1. ``make ec2`` This will create a small (4-node) k8s cluster on AWS EC2.
2. ``make run`` This runs the ray-operator and creates the example CRs from /deploy/crds.
3. ``make stop`` and ``make delete-ec2`` to delete ray-operator and teardown the cluster. 

Before running, make sure [kops](https://github.com/kubernetes/kops) and AWS Cli are properly installed. You may skip 1. if you already have a running k8s cluster and visible at .kubeconfig.