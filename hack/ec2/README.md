## Cluster

Before running the cluster, make sure all packages listed in `Preparation` section are exists, then you may need to:

1. change the cluster id in `build/kube/scripts/env.exp`
2. update file name of `build/kube/scripts/jxc-a3kvc.k8s.local.yaml` correspond to the id (replace "jxc-a3kvc" with the id)
3. update the spec.configBase field in the same file with the new id
4. update all field which contains "jxc-a3kvc" in the same file with the new id



First, run `python -m build.kube.cluster prepare` to create the state stores. The reason of doing these is due to the global unique name of AWS S3 bucket.

The binary store is not necessary unless you want to use a customized version of k8s.



#### Build up clusters

- To start a cluster, run `python -m build.kube.cluster up`
- To bring down a cluster, run `python -m build.kube.cluster down`

Updates to the scripts are welcome :D



## Preparation

#### Deploy Docker

Use command `docker version` to check whether Docker is installed, otherwise, follow [official installation guide](https://docs.docker.com/install/linux/docker-ce/ubuntu/):

```bash
sudo apt-get update
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo apt-key fingerprint 0EBFCD88
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

#### Deploy Kubernetes

Use command `kubectl version` to check whether Kubernetes is installed, otherwise, follow [official installation guide](https://kubernetes.io/docs/setup/independent/install-kubeadm/):

```bash
sudo apt-get update && apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

#### Deploy AWS-CLI and Config AWS Credentials

Use command `aws --version` to check whether aws-cli is installed, otherwise, follow:

```bash
sudo apt-get install aws-cli
```

Then, you should be sure AWS can be accessed with your credentials (Access Key and Secret Access)

Command `aws configure` can help you set your "AWS Access Key ID" and "AWS Secret Access Key"



#### Deploy metrics-server

Make sure that there is metrics-server in the path "./jxit/jxcore/hack/ec2/build", otherwise, clone it:

```bash
git clone https://github.com/kubernetes-incubator/metrics-server
```



#### Deploy helm

Use command `helm version` to check whether helm is installed, otherwise, follow:

```bash
mkdir -pv helm && cd helm
wget https://storage.googleapis.com/kubernetes-helm/helm-v2.9.1-linux-amd64.tar.gz
tar xf helm-v2.9.1-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin
rm -rf linux-amd64
```



## Connection

After creating a cluster on AWS ec2, you can connect it with the following command:

```bash
ssh -i $HOME/.ssh/id_rsa admin@ec2-34-215-180-148.us-west-2.compute.amazonaws.com
```

(34-215-180-148 is the ec2 IP, us-west-2 is region name.)



You can use the command below to view your ec2 IP:

```bash
aws ec2 describe-instances --query "Reservations[*].Instances[*].PublicIpAddress" --output text
```

Of course, you should configure your AWS before viewing, then input your ID, key, and region name(us-west-2, for example):

```bash
aws configure
```
