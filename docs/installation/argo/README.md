## Argo Workflows Installation
This section install Argo Workflows, follow the next for this:
1. Create a namespace called argo to install Argo Workflows
```
kubectl create ns argo
```
2. Install Argo Workflows using kubectl
```
kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo/stable/manifests/install.yaml
```
3. Because you are using k3s you have to support containerd as the container runtime with the next command:
```
kubectl patch configmap/workflow-controller-configmap \
-n argo \
--type merge \
-p '{"data":{"containerRuntimeExecutor":"k8sapi"}}'
```
4. Check that everything is running with the next command:
```
kubectl get pods -n argo
```
5. Access your Argo Workflow Deployment with port forward:
```
kubectl -n argo port-forward svc/argo-server 2746:2746
```
6. Access Argo Workflow on your browser accessing the next url:
```
http://127.0.0.1:2746
```
Note: If you are using port-forward to access Argo Workflows locally,  allow insecure connections from localhost in your browser. In Chrome, browse to: chrome://flags/. Search for “insecure” and you should see the option to “Allow invalid certificates for resources loaded from localhost.” Enable that option and restart your browser. Remember that by defaul Argo Workflows is installed with TLS.



## ArgoCD Installation
This section is to install ArgoCD with the next commands:
1. Create a namespace for ArgoCD:
```
kubectl create namespace argocd
```
2. Install ArgoCD using kubectl
```
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```
3. Create the ingress controller modifing the file inside the argocd folder called argocd-ingress.yaml with your desired domain, for that check the host and hosts sections inside the file, then apply the YAML file with the next command:
```
kubectl apply -f argocd/argocd-ingress.yaml
```
4. Set an A DNS record pointing to the subdomain where ArgoCD will be accesible
Note: Because this is one node Kubernetes, the IP of the node is the same IP for the Load Balancer

### ArgoCD Password
1. To get the ArgoCD password and generate a Token to launch ArgoCD get the argocd-server pod name, this will be the password to access ArgoCD, execute the next line to get argocd-server pod name:
```
kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-server -o name | cut -d'/' -f 2
```
2. Set a variable with the domain where ArgoCD is accesible
```
ARGOCD_SERVER=YourDomain
```
3. Generate the token to access the ArgoCD API, this is necessary to call ArgoCD when Argo Workflow need it
```
curl -sSL -k $ARGOCD_SERVER/api/v1/session -d $'{"username":"admin","password":"argocd-server-XXX-YYY"}'
```
Note: The password is the name of your argocd pod inside your argocd namespace
