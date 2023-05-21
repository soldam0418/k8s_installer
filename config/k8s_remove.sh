#/bin/bash
sudo kubeadm reset --force
sudo systemctl stop kubelet
sudo systemctl stop docker
sudo apt-get purge kubeadm kubelet kubectl docker --auto-remove -y
sudo rm -rf /var/lib/cni/
sudo rm -rf /var/lib/kubelet/*
sudo rm -rf /etc/cni/
sudo rm -rf /etc/kubernetes
sudo rm -rf $HOME/.kube
sudo ifconfig cni0 down
sudo ifconfig flannel.1 down
sudo ifconfig docker0 down