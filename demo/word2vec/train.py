import math
import pickle
import glob
import os
import sys
import paddle.v2 as paddle
import paddle.v2.dataset.common as common

embsize = 32
hiddensize = 256
N = 5

# NOTE: You need to generate and split dataset then put it under your cloud storage.
#       then you can use different size of embedding.

# NOTE: must change this to your own username on paddlecloud.
USERNAME = "wuyi05@baidu.com"
DC = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")
common.DATA_HOME = "/pfs/%s/home/%s" % (DC, USERNAME)
TRAIN_FILES_PATH = os.path.join(common.DATA_HOME, "imikolov")
WORD_DICT_PATH = os.path.join(common.DATA_HOME, "imikolov/word_dict.pickle")

TRAINER_ID = int(os.getenv("PADDLE_INIT_TRAINER_ID", "-1"))
TRAINER_COUNT = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS", "-1"))

def prepare_dataset():
    word_dict = paddle.dataset.imikolov.build_dict()
    with open(WORD_DICT_PATH, "w") as fn:
        pickle.dump(word_dict, fn)
    # convert will also split the dataset by line-count
    common.convert(TRAIN_FILES_PATH, 
                   paddle.dataset.imikolov.train(word_dict, N),
                   1000, "train")

def cluster_reader_recordio(trainer_id, trainer_count):
    '''
        read from cloud dataset which is stored as recordio format
        each trainer will read a subset of files of the whole dataset.
    '''
    import recordio
    def reader():
        TRAIN_FILES_PATTERN = os.path.join(TRAIN_FILES_PATH, "train-*")
        file_list = glob.glob(TRAIN_FILES_PATTERN)
        file_list.sort()
        my_file_list = []
        # read files for current trainer_id
        for idx, f in enumerate(file_list):
            if idx % trainer_count == trainer_id:
                my_file_list.append(f)
        for f in my_file_list:
            print "processing ", f
            reader = recordio.reader(f)
            record_raw = reader.read()
            while record_raw:
                yield pickle.loads(record_raw)
                record_raw = reader.read()
            reader.close()
    return reader

def cluster_reader_recordio_from_master(etcd_endpoints):
    '''
        call paddle master's RPC to get recordio metadata,
        then read from cloud storage.
    '''
    TRAIN_FILES_PATTERN = os.path.join(TRAIN_FILES_PATH, "train-*")
    return paddle.reader.creator.cloud_reader(TRAIN_FILES_PATTERN, etcd_endpoints)


def wordemb(inlayer):
    wordemb = paddle.layer.table_projection(
        input=inlayer,
        size=embsize,
        param_attr=paddle.attr.Param(
            name="_proj",
            initial_std=0.001,
            learning_rate=1,
            l2_rate=0, ))
    return wordemb


def main():
    paddle.init(use_gpu=False, trainer_count=1)
    # load dict from cloud file
    with open(WORD_DICT_PATH) as fn:
        word_dict = pickle.load(fn)
    dict_size = len(word_dict)
    firstword = paddle.layer.data(
        name="firstw", type=paddle.data_type.integer_value(dict_size))
    secondword = paddle.layer.data(
        name="secondw", type=paddle.data_type.integer_value(dict_size))
    thirdword = paddle.layer.data(
        name="thirdw", type=paddle.data_type.integer_value(dict_size))
    fourthword = paddle.layer.data(
        name="fourthw", type=paddle.data_type.integer_value(dict_size))
    nextword = paddle.layer.data(
        name="fifthw", type=paddle.data_type.integer_value(dict_size))

    Efirst = wordemb(firstword)
    Esecond = wordemb(secondword)
    Ethird = wordemb(thirdword)
    Efourth = wordemb(fourthword)

    contextemb = paddle.layer.concat(input=[Efirst, Esecond, Ethird, Efourth])
    hidden1 = paddle.layer.fc(
        input=contextemb,
        size=hiddensize,
        act=paddle.activation.Sigmoid(),
        layer_attr=paddle.attr.Extra(drop_rate=0.5),
        bias_attr=paddle.attr.Param(learning_rate=2),
        param_attr=paddle.attr.Param(
            initial_std=1. / math.sqrt(embsize * 8), learning_rate=1))
    predictword = paddle.layer.fc(
        input=hidden1,
        size=dict_size,
        bias_attr=paddle.attr.Param(learning_rate=2),
        act=paddle.activation.Softmax())

    def event_handler(event):
        if isinstance(event, paddle.event.EndIteration):
            if event.batch_id % 100 == 0:
                result = trainer.test(
                    paddle.batch(
                        # NOTE: if you're going to use cluster test files,
                        #       prepare them on the storage first
                        paddle.dataset.imikolov.test(word_dict, N), 32))
                print "Pass %d, Batch %d, Cost %f, %s, Testing metrics %s" % (
                    event.pass_id, event.batch_id, event.cost, event.metrics,
                    result.metrics)

    cost = paddle.layer.classification_cost(input=predictword, label=nextword)
    parameters = paddle.parameters.create(cost)
    adam_optimizer = paddle.optimizer.Adam(
        learning_rate=3e-3,
        regularization=paddle.optimizer.L2Regularization(8e-4))
    trainer = paddle.trainer.SGD(cost, parameters, adam_optimizer)
    trainer.train(
        # NOTE: use either cluster_reader_recordio or cluster_reader_recordio_from_master
        paddle.batch(cluster_reader_recordio(TRAINER_ID, TRAINER_COUNT), 32),
        num_passes=30,
        event_handler=event_handler)


if __name__ == '__main__':
    usage = "python train.py [prepare|train]"
    if len(sys.argv) != 2:
        print usage
        exit(1)

    if TRAINER_ID == -1 or TRAINER_COUNT == -1:
        print "no cloud environ found, must run on cloud"
        exit(1)

    if sys.argv[1] == "prepare":
        prepare_dataset()
    elif sys.argv[1] == "train":
        main()
