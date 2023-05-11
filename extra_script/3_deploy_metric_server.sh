#/bin/bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.5.1/components.yaml
# v0.5.1
kubectl get deploy metrics-server -n kube-system -o yaml > metrics-server.yaml
sudo sed -i'' -r -e "/- args:/a\        - --kubelet-insecure-tls" metrics-server.yaml
kubectl replace -f metrics-server.yaml --force