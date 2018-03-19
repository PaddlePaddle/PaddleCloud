# PIPE_READER 读取HDFS数据指南
> pipe_reader 是以数据流的形式读取数据，然后用定义好的parser将数据处理成所需的格式，通过Python的yield进行数据的返回

## 使用方法
### 数据准备
1. 按照集群版本的数据准备方法进行[数据准备](https://github.com/PaddlePaddle/cloud/blob/develop/doc/usage_cn.md#%E5%87%86%E5%A4%87%E8%AE%AD%E7%BB%83%E6%95%B0%E6%8D%AE)
2. 准备数据文件用于控制集群任务的并发度，**此文件的数量与1中准备的要一致**
在集群上进行训练时，数据在机器之间的分配以文件为最小粒度，paddle的集群版会根据文件的数量来决定启动多少个节点来执行训练。padlle当前对数据的读取方案是将用户指定的数据按照文件分别下载集群中的节点上，然后进行本地读取，但我们采用pipe_reader可直接从HDFS上以数据流的形式进行数据读取，因此我们需要准备两份训练数据数据。一份用于指定给paddle来分配集群执行节点的数据（此份数据会被paddle下载到节点本地），另外一份是实际的训练数据，将交给pipe_reader进行读取。
训练数据目录举例，train1的目录为需要指定给paddle mpi中训练目录的参数用于分配节点数量，文件数量跟train2保持一直，文件可为空，或有少量数据：
train1中part-*
```
/paddle/cluster_demo/text_classification/data/train1/part-00000
/paddle/cluster_demo/text_classification/data/train1/part-00001
/paddle/cluster_demo/text_classification/data/train1/part-00002
```
训练数据目录举例，train2的目录为需要指定给pipe_reader进行数据的读取的数据，是实际的训练数据：
```
/paddle/cluster_demo/text_classification/data/train2/part-00000
/paddle/cluster_demo/text_classification/data/train2/part-00001
/paddle/cluster_demo/text_classification/data/train2/part-00002
```
### 代码准备
1. 添加pipe_reader实现到代码中
**由于集群上paddle版本的原因**，暂时没有paddle的集群版中没有pipe_reader的实现，可自行在代码中将pipe_reader的实现加入
需要加入**自己的**代码中的pipe_reader代码，代码见**api_train_pipe_reader.py**

2. 指定解析的parser
pipe_reader将解析出的多行数据交由parser进行解析，将数据解析成所需要的格式
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
3. 指定pipe_reader参数
读取文本格式的HDFS文件，只需要指定pipe_reader的还需要指定left_cmd用于读取数据流。集群中每个节点都会被指定一个node_id，依次从0开始，比如有3份训练文件的node_id分别是 0，1，2。
因此通过node_id，就可以区分出当前执行节点需要读取的数据。
```
/paddle/cluster_demo/text_classification/data/train2/part-00000
/paddle/cluster_demo/text_classification/data/train2/part-00001
/paddle/cluster_demo/text_classification/data/train2/part-00002
```
在已经知道了node_id的情况下，根据代码中自行定义好的HDFS路径前缀，进行拼接，就可以获取到最终的HDFS路径，MPI集群上已经有HADOOP客户端了，但是环境复杂，**在使用HDFS数据的时候，请带上HDFS的集群前缀或者再hadoop的参数中指定正确的fs.default.name**
举例一个left_cmd：
```
hadoop fs -Dfs.default.name=hdfs://hadoop.com/:54310 -Dhadoop.job.ugi=name,password -cat /paddle/cluster_demo/text_classification/data/train/part-00000
```
4. 指定train的reader
将pipe_reader 指定为train/test的reader或者指定为paddle.batch的参数，
例如：
```python
trainer.train(
        paddle.batch(pipe_reader(gene_cmd(int(node_id), "train"), uni_parser), 32),
        num_passes=30,
        event_handler=event_handler)
```
### 提交集群训练任务
同[通过Receiver提交集群训练任务](https://github.com/PaddlePaddle/cloud/blob/develop/doc/usage_cn.md#%E6%8F%90%E4%BA%A4%E4%BB%BB%E5%8A%A1)
**注意：**
--train_data_path ${HDFS_train_path} 此参数需要指定的是用于指定集群执行数量的数据目录


