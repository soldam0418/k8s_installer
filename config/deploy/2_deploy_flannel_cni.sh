#/bin/bash
sudo apt-get install wget -y
#sudo wget https://raw.githubusercontent.com/flannel-io/flannel/v0.15.0/Documentation/kube-flannel.yml
#sudo sed -i -e 's?10.244.0.0/16? 192.168.0.0/16?g' kube-flannel.yml
#sudo kubectl apply -f kube-flannel.yml
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/v0.15.0/Documentation/kube-flannel.yml
# v1.15.0 / Source : https://github.com/flannel-io/flannel/blob/2d8456cb2e8fe3d841a46f331771d4d2cf07a4e5/Documentation/kube-flannel.yml
