# Copyright (c) 2016 PaddlePaddle Authors. All Rights Reserved
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import paddle.v2 as paddle
import paddle.v2.dataset.common as common
import os
import sys
import glob
import pickle

# NOTE: must change this to your own username on paddlecloud.
USERNAME = "demo"
DC = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")
common.DATA_HOME = "/pfs/%s/home/%s" % (DC, USERNAME)
TRAIN_FILES_PATH = os.path.join(common.DATA_HOME, "imdb")
TEST_FILES_PATH = os.path.join(common.DATA_HOME, "imdb")

TRAINER_ID = int(os.getenv("PADDLE_INIT_TRAINER_ID", "-1"))
TRAINER_COUNT = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS", "-1"))

def prepare_dataset():
    word_dict = paddle.dataset.imdb.word_dict()
    # convert will also split the dataset by line-count
    common.convert(TRAIN_FILES_PATH,
            lambda: paddle.dataset.imdb.train(word_dict),
                1000, "train")
    common.convert(TEST_FILES_PATH,
            lambda: paddle.dataset.imdb.test(word_dict),
                1000, "test")

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



def convolution_net(input_dim, class_dim=2, emb_dim=128, hid_dim=128):
    data = paddle.layer.data("word",
                             paddle.data_type.integer_value_sequence(input_dim))
    emb = paddle.layer.embedding(input=data, size=emb_dim)
    conv_3 = paddle.networks.sequence_conv_pool(
        input=emb, context_len=3, hidden_size=hid_dim)
    conv_4 = paddle.networks.sequence_conv_pool(
        input=emb, context_len=4, hidden_size=hid_dim)
    output = paddle.layer.fc(
        input=[conv_3, conv_4], size=class_dim, act=paddle.activation.Softmax())
    lbl = paddle.layer.data("label", paddle.data_type.integer_value(2))
    cost = paddle.layer.classification_cost(input=output, label=lbl)
    return cost


def stacked_lstm_net(input_dim,
                     class_dim=2,
                     emb_dim=128,
                     hid_dim=512,
                     stacked_num=3):
    """
    A Wrapper for sentiment classification task.
    This network uses bi-directional recurrent network,
    consisting three LSTM layers. This configure is referred to
    the paper as following url, but use fewer layrs.
        http://www.aclweb.org/anthology/P15-1109

    input_dim: here is word dictionary dimension.
    class_dim: number of categories.
    emb_dim: dimension of word embedding.
    hid_dim: dimension of hidden layer.
    stacked_num: number of stacked lstm-hidden layer.
    """
    assert stacked_num % 2 == 1

    layer_attr = paddle.attr.Extra(drop_rate=0.5)
    fc_para_attr = paddle.attr.Param(learning_rate=1e-3)
    lstm_para_attr = paddle.attr.Param(initial_std=0., learning_rate=1.)
    para_attr = [fc_para_attr, lstm_para_attr]
    bias_attr = paddle.attr.Param(initial_std=0., l2_rate=0.)
    relu = paddle.activation.Relu()
    linear = paddle.activation.Linear()

    data = paddle.layer.data("word",
                             paddle.data_type.integer_value_sequence(input_dim))
    emb = paddle.layer.embedding(input=data, size=emb_dim)

    fc1 = paddle.layer.fc(
        input=emb, size=hid_dim, act=linear, bias_attr=bias_attr)
    lstm1 = paddle.layer.lstmemory(
        input=fc1, act=relu, bias_attr=bias_attr, layer_attr=layer_attr)

    inputs = [fc1, lstm1]
    for i in range(2, stacked_num + 1):
        fc = paddle.layer.fc(
            input=inputs,
            size=hid_dim,
            act=linear,
            param_attr=para_attr,
            bias_attr=bias_attr)
        lstm = paddle.layer.lstmemory(
            input=fc,
            reverse=(i % 2) == 0,
            act=relu,
            bias_attr=bias_attr,
            layer_attr=layer_attr)
        inputs = [fc, lstm]

    fc_last = paddle.layer.pooling(
        input=inputs[0], pooling_type=paddle.pooling.Max())
    lstm_last = paddle.layer.pooling(
        input=inputs[1], pooling_type=paddle.pooling.Max())
    output = paddle.layer.fc(
        input=[fc_last, lstm_last],
        size=class_dim,
        act=paddle.activation.Softmax(),
        bias_attr=bias_attr,
        param_attr=para_attr)

    lbl = paddle.layer.data("label", paddle.data_type.integer_value(2))
    cost = paddle.layer.classification_cost(input=output, label=lbl)
    return cost


def main():
    # init
    paddle.init()
    #data
    print 'load dictionary...'
    word_dict = paddle.dataset.imdb.word_dict()
    dict_dim = len(word_dict)
    class_dim = 2
    train_reader = paddle.batch(
        paddle.reader.shuffle(
            cluster_reader_recordio(TRAINER_ID, TRAINER_COUNT, "train"), buf_size=1000),
        batch_size=100)
    test_reader = paddle.batch(
        cluster_reader_recordio(TRAINER_ID, TRAINER_COUNT, "test"), batch_size=100)

    feeding = {'word': 0, 'label': 1}

    # network config
    # Please choose the way to build the network
    # by uncommenting the corresponding line.
    cost = convolution_net(dict_dim, class_dim=class_dim)
    # cost = stacked_lstm_net(dict_dim, class_dim=class_dim, stacked_num=3)

    # create parameters
    parameters = paddle.parameters.create(cost)

    # create optimizer
    adam_optimizer = paddle.optimizer.Adam(
        learning_rate=2e-3,
        regularization=paddle.optimizer.L2Regularization(rate=8e-4),
        model_average=paddle.optimizer.ModelAverage(average_window=0.5))

    # End batch and end pass event handler
    def event_handler(event):
        if isinstance(event, paddle.event.EndIteration):
            if event.batch_id % 100 == 0:
                print "\nPass %d, Batch %d, Cost %f, %s" % (
                    event.pass_id, event.batch_id, event.cost, event.metrics)
            else:
                sys.stdout.write('.')
                sys.stdout.flush()
        if isinstance(event, paddle.event.EndPass):
            result = trainer.test(reader=test_reader, feeding=feeding)
            print "\nTest with Pass %d, %s" % (event.pass_id, result.metrics)

    # create trainer
    trainer = paddle.trainer.SGD(
        cost=cost, parameters=parameters, update_equation=adam_optimizer)

    trainer.train(
        reader=train_reader,
        event_handler=event_handler,
        feeding=feeding,
        num_passes=2)

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
