## Preface
We convert the common datasets such as `cifar` `imdb` and so on to recordio format, and then we can start a distribute train jobs more convenient after that.

## Build Docker image
```
docker build . -t paddlepaddle/recordiodataset:latest
docker push paddlepaddle/recordiodataset:latest
```

## Run on kubernetes cluster

This command will convert dataset's format to recordio.

```
kubectrl create -f convert_app.yaml
```