from PIL import Image
import numpy as np
import paddle.v2 as paddle
import paddle.v2.dataset.common as common
from paddle.v2.reader.creator import cloud_reader
import os
import sys
import glob
import pickle


DC = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")

DATASET_PATH = "/data/mnist/mnist-train-*"

def softmax_regression(img):
    predict = paddle.layer.fc(
        input=img, size=10, act=paddle.activation.Softmax())
    return predict
'''
pass_num = 0
def cloud_reader(paths, etcd_endpoints, timeout_sec=5, buf_size=64):
    """
    Create a data reader that yield a record one by one from
        the paths:
    :paths: path of recordio files, can be a string or a string list.
    :etcd_endpoints: the endpoints for etcd cluster
    :returns: data reader of recordio files.

    ..  code-block:: python
        from paddle.v2.reader.creator import cloud_reader
        etcd_endpoints = "http://127.0.0.1:2379"
        trainer.train.(
            reader=cloud_reader(["/work/dataset/uci_housing/uci_housing*"], etcd_endpoints),
        )
    """
    import os
    import cPickle as pickle
    import paddle.v2.master as master
    c = master.client(etcd_endpoints, timeout_sec, buf_size)

    if isinstance(paths, basestring):
        path = [paths]
    else:
        path = paths
    c.set_dataset(path)

    def reader():
        global pass_num
        c.paddle_start_get_records(pass_num)
        pass_num += 1
        print 'call reader'
        while True:
            print 'before next record'
            r, e = c.next_record()
            print 'next record', r, e
            if not r:
                if e != -2:
                    print "get record error: ", e
                break
            yield pickle.loads(r)

    return reader
'''
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
    etcd_ip = os.getenv("ETCD_IP")
    etcd_endpoint = "http://" + etcd_ip + ":" + "2379"
    paddle.init(trainer_count=1)

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
        cost=cost,
        parameters=parameters,
        update_equation=optimizer,
        is_local=False,
        pserver_spec=etcd_endpoint,
        use_etcd=True)

    def event_handler(event):
        if isinstance(event, paddle.event.EndIteration):
            if event.batch_id % 100 == 0:
                print "Pass %d, Batch %d, Cost %f, %s" % (
                    event.pass_id, event.batch_id, event.cost, event.metrics)
        if isinstance(event, paddle.event.EndPass):
            result = trainer.test(
                    reader=paddle.batch(
                        paddle.dataset.mnist.test(),
                    batch_size=2))
            print "Test with Pass %d, Cost %f, %s\n" % (
                event.pass_id, result.cost, result.metrics)

    trainer.train(
        reader=paddle.batch(
            cloud_reader(
                [DATASET_PATH], etcd_endpoint),
            batch_size=10),
        event_handler=event_handler,
        num_passes=10)

if __name__ == '__main__':
    main()
