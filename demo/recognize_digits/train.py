from PIL import Image
import numpy as np
import paddle.v2 as paddle
import paddle.v2.dataset.common as common
import os
import sys
import glob
import pickle


# NOTE: must change this to your own username on paddlecloud.
USERNAME = "wanghaoshuang@baidu.com"
DC = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")
common.DATA_HOME = "/pfs/%s/home/%s" % (DC, USERNAME)
TRAIN_FILES_PATH = os.path.join(common.DATA_HOME, "mnist")
TEST_FILES_PATH = os.path.join(common.DATA_HOME, "mnist")

TRAINER_ID = int(os.getenv("PADDLE_INIT_TRAINER_ID", "-1"))
TRAINER_COUNT = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS", "-1"))

def prepare_dataset():
    # convert will also split the dataset by line-count
    common.convert(TRAIN_FILES_PATH,
                paddle.dataset.mnist.train(),
                8192, "train")
    common.convert(TEST_FILES_PATH,
                paddle.dataset.mnist.test(),
                1, "test")

def cluster_reader_recordio(trainer_id, trainer_count, flag):
    '''
        read from cloud dataset which is stored as recordio format
        each trainer will read a subset of files of the whole dataset.
    '''
    import recordio
    def reader():
        PATTERN_STR = "%s-*" % flag
        FILES_PATTERN = os.path.join(TRAIN_FILES_PATH, PATTERN_STR)
        file_list = glob.glob(FILES_PATTERN)
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



def softmax_regression(img):
    predict = paddle.layer.fc(
        input=img, size=10, act=paddle.activation.Softmax())
    return predict


def multilayer_perceptron(img):
    # The first fully-connected layer
    hidden1 = paddle.layer.fc(input=img, size=128, act=paddle.activation.Relu())
    # The second fully-connected layer and the according activation function
    hidden2 = paddle.layer.fc(
        input=hidden1, size=64, act=paddle.activation.Relu())
    # The thrid fully-connected layer, note that the hidden size should be 10,
    # which is the number of unique digits
    predict = paddle.layer.fc(
        input=hidden2, size=10, act=paddle.activation.Softmax())
    return predict


def convolutional_neural_network(img):
    # first conv layer
    conv_pool_1 = paddle.networks.simple_img_conv_pool(
        input=img,
        filter_size=5,
        num_filters=20,
        num_channel=1,
        pool_size=2,
        pool_stride=2,
        act=paddle.activation.Relu())
    # second conv layer
    conv_pool_2 = paddle.networks.simple_img_conv_pool(
        input=conv_pool_1,
        filter_size=5,
        num_filters=50,
        num_channel=20,
        pool_size=2,
        pool_stride=2,
        act=paddle.activation.Relu())
    # fully-connected layer
    predict = paddle.layer.fc(
        input=conv_pool_2, size=10, act=paddle.activation.Softmax())
    return predict


def main():
    paddle.init()

    # define network topology
    images = paddle.layer.data(
        name='pixel', type=paddle.data_type.dense_vector(784))
    label = paddle.layer.data(
        name='label', type=paddle.data_type.integer_value(10))

    # Here we can build the prediction network in different ways. Please
    # choose one by uncomment corresponding line.
    # predict = softmax_regression(images)
    # predict = multilayer_perceptron(images)
    predict = convolutional_neural_network(images)

    cost = paddle.layer.classification_cost(input=predict, label=label)

    parameters = paddle.parameters.create(cost)

    optimizer = paddle.optimizer.Momentum(
        learning_rate=0.1 / 128.0,
        momentum=0.9,
        regularization=paddle.optimizer.L2Regularization(rate=0.0005 * 128))

    trainer = paddle.trainer.SGD(
        cost=cost, parameters=parameters, update_equation=optimizer)

    def event_handler(event):
        if isinstance(event, paddle.event.EndIteration):
            if event.batch_id % 100 == 0:
                print "Pass %d, Batch %d, Cost %f, %s" % (
                    event.pass_id, event.batch_id, event.cost, event.metrics)
        if isinstance(event, paddle.event.EndPass):
            result = trainer.test(
                    reader=paddle.batch(
                    cluster_reader_recordio(TRAINER_ID, TRAINER_COUNT, "test"),
                    batch_size=2))
            print "Test with Pass %d, Cost %f, %s\n" % (
                event.pass_id, result.cost, result.metrics)

    trainer.train(
        reader=paddle.batch(
            cluster_reader_recordio(TRAINER_ID, TRAINER_COUNT, "train"),
            batch_size=128),
        event_handler=event_handler,
        num_passes=5)

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
