# PipeReader 读取HDFS数据指南
PipeReader 是以数据流的形式读取数据，然后用定义好的`parser`将数据处理成所需的格式，通过Python的yield进行数据的返回

## 使用方法
### 数据准备
1. 按照集群版本的数据准备方法进行[数据准备](https://github.com/PaddlePaddle/cloud/blob/develop/doc/usage_cn.md#%E5%87%86%E5%A4%87%E8%AE%AD%E7%BB%83%E6%95%B0%E6%8D%AE)

### 代码准备
1. 添加PipeReader实现到代码中
[PipeReader](https://github.com/PaddlePaddle/Paddle/tree/b4302bbbb85bbfd984cb2825887c133120699775/python/paddle/v2/reader/decorator.py)的代码已合入Paddle代码库中，如果您使用的Paddle镜像不包含此代码，可以通过更新镜像的方式来升级paddle

2. 指定解析的parser
PipeReader将解析出的多行数据交由parser进行解析，将数据解析成所需要的格式
一个简单的parser:
```python
    def sample_parser(lines):
        # parse each line as one sample data,
        # return a list of samples as batches.
        ret = []
        for l in lines:
            ret.append(l.split(" ")[1:5])
        return ret
```

3. PipeReader的`command`参数说明
通过合适的`command`参数指定，可以从HDFS/Ceph/URL/FTP/AWS 等很多目的地址读取数据，例如：
读取HDFS数据： `command = "hadoop fs -cat /path/to/some/file"`
读取本地文件：   `command = "cat sample_file.tar.gz"`
读取HTTP地址：`command = "curl http://someurl"`
读取其他程序的标准输出： `command = "python print_s3_bucket.py"`

4. PipeReader读取HDFS数据的`command`参数说明
如需要读取文本格式的HDFS文件，则可以指定PipeReader的`command`参数为`hadoop fs -cat /path/to/some/file`。集群中每个节点都会被指定一个node_id，在PaddleCloud环境下，可以通过获取**环境变量 PADDLE_INIT_TRAINER_ID** 来获取当前执行任务的节点的node_id，依次从0开始，比如有3份训练文件的node_id分别是 0，1，2。
因此通过node_id，每个训练节点就可以区分出当前执行节点所需要读取的数据。
```
/paddle/cluster_demo/text_classification/data/train2/part-00000
/paddle/cluster_demo/text_classification/data/train2/part-00001
/paddle/cluster_demo/text_classification/data/train2/part-00002
```
在已经知道了node_id的情况下，根据代码中自行定义好的HDFS路径前缀，进行拼接，就可以获取到最终的HDFS路径，PaddleCloud集群上有集成好HADOOP客户端的镜像，使用的时候可以根据数据所在的集群地址进行相应的参数替换：
**在使用HDFS数据的时候，请带上HDFS的集群前缀或者再hadoop的参数中指定正确的fs.default.name和hadoop.job.ugi等参数**
举一个`command`参数的例子：
```
hadoop fs -Dfs.default.name=hdfs://hadoop.com/:54310 -Dhadoop.job.ugi=name,password -cat /paddle/cluster_demo/text_classification/data/train/part-00000
```

5. 在您的代码中使用`PipeReader`
您可以将`PipeReader`指定为`trainer.train`, `trainer.test` 或作为`paddle.batch`的参数，例如：
```python
trainer.train(
        paddle.batch(PipeReader(gene_cmd(int(node_id), "train")), 32),
        num_passes=30,
        event_handler=event_handler)
```
或者，您也可以将`PipeReader`作为其他`Reader`的参数，对`PipeReader`输出的数据做二次处理，例如：
```python
def example_reader():
    for f in myfiles:
        pr = PipeReader("cat %s"%f)
        for l in pr.get_line():
            sample = l.split(" ")
            yield sample

trainer.train(
        paddle.batch(example_reader, 32),
        num_passes=30,
        event_handler=event_handler)
```