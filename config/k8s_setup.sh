#/bin/bash
sudo swapoff -a && sed -i '/swap/s/&/#/' /etc/fstab
sudo systemctl stop firewalld && sudo systemctl disable firewalld
sudo curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo mkdir /etc/docker
sudo cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
sudo systemctl enable docker
sudo systemctl start docker
sudo sed -i -e '/disabled_plugins/ s/^/#/' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
sudo yum update
sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
sudo systemctl daemon-reload
sudo systemctl restart kubelet