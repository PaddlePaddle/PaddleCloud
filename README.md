# PaddlePaddle Cloud

PaddlePaddle Cloud is a Distributed Deep-Learning Cloud Platform for both cloud
providers and enterprises.

PaddlePaddle Cloud use [Kubernetes](https://kubernetes.io) as it's backend job
dispatching and cluster resource management center. And use [PaddlePaddle](https://github.com/PaddlePaddle/Paddle.git)
as the deep-learning frame work. Users can use web pages or command-line tools
to submit their deep-learning training jobs remotely to make use of power of
large scale GPU clusters.

## Using Command-line To Submit Cloud Training Jobs

[中文手册](./doc/usage_cn.md)

English tutorials(comming soon...)

## Deploy PaddlePaddle Cloud

### Pre-Requirements
- PaddlePaddle Cloud use kubernetes as it's backend core, deploy kubernetes cluster
  using [Sextant](https://github.com/k8sp/sextant) or any tool you like.


### Run on kubernetes
- Build Paddle Cloud Docker Image

  ```bash
  # build docker image
  git clone https://github.com/PaddlePaddle/cloud.git
  cd cloud/paddlecloud
  docker build -t [your_docker_registry]/pcloud .
  # push to registry so that we can submit paddlecloud to kubernetes
  docker push [your_docker_registry]/pcloud
  ```

- We use [volume](https://kubernetes.io/docs/concepts/storage/volumes/) to mount MySQL data,
  cert files and settings, in `k8s/` folder we have some samples for how to mount
  stand-alone files and settings using [hostpath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath). Here's
  a good tutorial of creating kubernetes certs: https://coreos.com/kubernetes/docs/latest/getting-started.html

  - create data folder on a Kubernetes node, such as:
  
  ```bash
  mkdir -p /home/pcloud/data/mysql
  mkdir -p /home/pcloud/data/certs
  ```
  - Copy Kubernetes CA files (ca.pem, ca-key.pem) to `/home/pcloud/data/certs` folder
  - Copy Kubernetes admin user key (admin.pem, admin-key.pem) to `/home/pcloud/data/certs` folder (if you don't have it on Kubernetes worker node, you'll find it on Kubernetes master node)
  - Optianal: copy CephFS Key file(admin.secret) to `/home/pcloud/data/certs` folder
  - Copy `paddlecloud/settings.py` file to `/home/pcloud/data` folder

- Configure `cloud_deployment.yaml`
  - `spec.template.spec.containers[0].volumes` change the `hostPath` which match your data folder.
  - `spec.template.spec.nodeSelector.`, edit the value `kubernetes.io/hostname` to host which data folder on.You can use `kubectl get nodes` to list all the Kubernetes nodes.
- Configure `settings.py`
  - Add your domain name (or the paddle cloud server's hostname or ip address) to `ALLOWED_HOSTS`.
  - Configure `DATACENTERS` to your backend storage, supports CephFS and HostPath currently.
    You can use HostPath mode to make use of shared file-systems like "NFS".
    If you use something like hostPath, you need to modify the DATACENTERS field in settings.py as follows:
   
    ```
    DATACENTERS = {
        "<MY_DATA_CENTER_NAME_HERE>":{
            "fstype": "hostpath",
            "host_path": "/home/pcloud/data/public/",
            "mount_path": "/pfs/%s/home/%s/" # mount_path % ( dc, username )
        }
    }
    ```
    
- Configure `cloud_ingress.yaml` is your kubernetes cluster is using [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) (if you need to use Jupyter notebook, you have to configure the ingress controller)
  to proxy HTTP traffics, or you can configure `cloud_service.yaml` to use [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport)
  - if using ingress, configure `spec.rules[0].host` to your domain name
- Deploy mysql on Kubernetes first if you don't have it on your cluster, and modify the mysql endpoint in settings.py
  - `kubectl create -f ./mysql_deployment.yaml` (you need to fill in the nodeselector field with your node's hostname or ip in yaml first)
  - `kubectl create -f ./mysql_service.yaml`
- Deploy cloud on Kubernetes
  - `kubectl create -f k8s/cloud_deployment.yaml`(you need to fill in the nodeselector field with your node's hostname or ip in yaml first)
  - `kubectl create -f k8s/cloud_service.yaml`
  - `kubectl create -f k8s/cloud_ingress.yaml`(optianal if you don't need Jupyter notebook)


To test or visit the website, find out the kubernetes ingress IP
addresses, or the NodePort.

Then open your browser and visit `http://<ingress-ip-address>`, or
`http://<any-node-ip-address>:<NodePort>`

- Prepare public dataset

  You can create a Kubernetes Job for preparing the public cloud dataset with RecordIO files. You should modify the YAML file as your environment:
  - `<DATACENTER>`, Your cluster datacenter 
  - `<MONITOR_ADDR>`, Ceph monitor address
  ```bash
  kubectl create -f k8s/prepare_dataset.yaml
  ```

### Run locally without docker

- You still need a kubernetes cluster when try running locally.
- Make sure you have `Python > 2.7.10` installed.
- Python needs to support `OPENSSL 1.2`. To check it out, simply run:
    ```python
       >>> import ssl
       >>> ssl.OPENSSL_VERSION
       'OpenSSL 1.0.2k  26 Jan 2017'
    ```
- Make sure you are using a virtual environment of some sort (e.g. `virtualenv` or
`pyenv`).

```
virtualenv paddlecloudenv
# enable the virtualenv
source paddlecloudenv/bin/activate
```

To run for the first time, you need to:

```
cd paddlecloud
npm install
pip install -r requirements.txt
./manage.py migrate
./manage.py loaddata sites
npm run dev
```

Browse to http://localhost:8000/

If you are starting the server for the second time, just run:
```
./manage.py runserver
```

### Configure Email Sending
If you want to use `mail` command to send confirmation emails, change the below settings:

```
EMAIL_BACKEND = 'django_sendmail_backend.backends.EmailBackend'
```

You may need to use `hostNetwork` for your pod when using mail command.

Or you can use django smtp bindings just refer to https://docs.djangoproject.com/en/1.11/topics/email/
