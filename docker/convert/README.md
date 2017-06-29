## 面向对象
集群管理者

## 目的
当我们有了新的集群或者需要重新生成一下数据，只需要改一下`convert_app.yaml`里边的路径，然后用`kubectrl create -f convert_app.yaml` 就可以把数据写到相应的位置。

## 构建docker image

```
docker build . -t paddlepaddle/recordiodataset:latest
docker push paddlepaddle/recordiodataset:latest
```

## 启动任务

```
kubectrl create -f convert_app.yaml
```