#/bin/bash
sudo kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.5.1/components.yaml
# v0.5.1
sudo kubectl get deploy metrics-server -n kube-system -o yaml > /home/metrics-server.yaml
sudo sed -i'' -r -e "/- args:/a\        - --kubelet-insecure-tls" /home/metrics-server.yaml
sudo kubectl replace -f /home/metrics-server.yaml --force