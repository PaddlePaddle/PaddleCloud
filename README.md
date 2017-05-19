# PaddlePaddle Cloud

## Getting Started

### Pre-Requirements
- PaddlePaddle Cloud needs python to support `OPENSSL 1.2`. To check it out, simply run:
    ```python
       >>> import ssl
       >>> ssl.OPENSSL_VERSION
       'OpenSSL 1.0.2k  26 Jan 2017'
    ```
- Make sure you have `Python > 2.7.10` installed.

### Run on kubernetes
- Build Paddle Cloud Docker Image
  ```bash
  # build docker image
  git clone https://github.com/PaddlePaddle/cloud.git
  cd cloud/paddlecloud
  docker build -t [your_docker_registry]/pcloud .
  docker push [your_docker_registry]/pcloud
  ```
- We use [volume](https://kubernetes.io/docs/concepts/storage/volumes/) to mount MySQL data and cert files, such as CephFS, GlusterFS and etc..., the follow is a example using [hostpath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath):

  - create data folder on a Kubernetes node, such as:
  ```bash
  mkdir -p /home/pcloud/data/mysql
  mkdir -p /home/pcloud/data/certs
  ```
  - Copy Kubernetes CA files (ca.pem, ca-key.pem, ca.srl) to `pcloud_data/certs` folder
  - Copy CephFS Key file(admin.secret) to `pcloud_data/certs` folder
  - Copy `/paddlecloud/settings.py` file to `pcloud_data` folder

- Configure `cloud_deployment.yaml`
  - `spec.template.spec.containers[0].volumes` change the `hostPath` which match your data folder.
  - `spec.template.spec.nodeSelector.`, edit the value `kubernetes.io/hostname` to host which data folder on.You can use `kubectl get nodes` to list all the Kubernetes nodes.
- Configure `settings.py`
  - Add your domain name to `ALLOWED_HOSTS`.
  - Configure `DATACENTERS` to your backend storage.
- Configure `cloud_ingress.yaml`
  - `spec.rules[0].host` specify your domain name
- Deploy cloud on Kubernetes
  - `kubectl create -f k8s/cloud_deployment.yaml`
  - `kubectl create -f k8s/cloud_service.yaml`
  - `kubectl create -f k8s/cloud_ingress.yaml`


To test or visit the website, find out the kubernetes [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) IP addresses, and bind it to your `/etc/hosts` file:
```
# your ingress IP address
192.168.1.100    cloud.paddlepaddle.org
```

Then open your browser and visit http://cloud.paddlepaddle.org.

### Run locally
Make sure you are using a virtual environment of some sort (e.g. `virtualenv` or
`pyenv`).
```
virtualenv paddlecloudenv
# enable the virtualenv
source paddlecloudenv/bin/activate
```

To run for the first time, you need to:
```
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
