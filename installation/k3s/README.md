## K3s installation

The following commands have to be executed inside your virtual machine:

1. First update your Ubuntu
```
sudo apt-get update
```
2. Set a variable with your Public IP
```
PUBLIC_IP=YOUR_IP
```
3. Install k3s with the next command
```
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik --tls-san "$PUBLIC_IP" --node-external-ip "$PUBLIC_IP" --write-kubeconfig-mode 644" sh -s -
```
4. Check that your unique node is on Ready status, with the next command
```
kubectl get nodes
```
5. Install helm with the following commands
```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```
6. Download the kubeconfig in your local machine
```
ssh -i id_rsa yourUser@yourDomain cat /etc/rancher/k3s/k3s.yaml > ~/.kube/config
```
7. Change the Kubernetes API connection from:  
server: https://127.0.0.1:6443  
to  
server: https://yourDomain:6443
