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
```bash
# build docker image
git clone https://github.com/PaddlePaddle/cloud.git
cd cloud/paddlecloud
docker build -t [your_docker_registry]/pcloud .
docker push [your_docker_registry]/pcloud
# submit to kubernetes
kubectl create -f ./k8s
```

To test or visit the web site, find out the kubernetes [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) ip addresses, and bind it to your `/etc/hosts` file:
```
# your ingress ip address
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

