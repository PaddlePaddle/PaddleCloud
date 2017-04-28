# PaddlePaddle Cloud

## Getting Started

Make sure you have `Python > 2.7.10` installed.

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
