import paddle.v2 as paddle
import pcloud.dataset.uci_housing as uci_housing
import os
import gzip
import glob
import recordio
import cPickle as pickle
trainer_id = int(os.getenv("PADDLE_INIT_TRAINER_ID"))
trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS"))

def recordio_reader_creator(path):
    files = glob.glob(path)
    files.sort()
    flie_count = len(files)
    files_current_train = []

    for idx, fn in enumerate(files): 
        if idx % trainer_count == trainer_id:
            files_current_train.append(fn)

    def reader():
        for fn in files_current_train:
            reader = recordio.reader(fn)
            while True:
                r = reader.read()
                if r is None:
                    break 
                yield pickle.loads(r) 
            reader.close()
    return reader

def main():
    # init
    paddle.init()

    # network config
    x = paddle.layer.data(name='x', type=paddle.data_type.dense_vector(13))
    y_predict = paddle.layer.fc(input=x, size=1, act=paddle.activation.Linear())
    y = paddle.layer.data(name='y', type=paddle.data_type.dense_vector(1))
    cost = paddle.layer.mse_cost(input=y_predict, label=y)

    # create parameters
    parameters = paddle.parameters.create(cost)

    # create optimizer
    optimizer = paddle.optimizer.Momentum(momentum=0)

    trainer = paddle.trainer.SGD(
        cost=cost, parameters=parameters, update_equation=optimizer, is_local=False)

    feeding = {'x': 0, 'y': 1}

    # event_handler to print training and testing info
    def event_handler(event):
        if isinstance(event, paddle.event.EndIteration):
            if event.batch_id % 100 == 0:
                print "Pass %d, Batch %d, Cost %f" % (
                    event.pass_id, event.batch_id, event.cost)

        if isinstance(event, paddle.event.EndPass):
            result = trainer.test(
                reader=paddle.batch(uci_housing.test(), batch_size=2),
                feeding=feeding)
            print "Test %d, Cost %f" % (event.pass_id, result.cost)
            if trainer_id == "0":
                with gzip.open("fit-a-line_pass_%05d.tar.gz" % event.pass_id,
                               "w") as f:
                    parameters.to_tar(f)
    # training
    trainer.train(
        reader=paddle.batch(
            paddle.reader.shuffle(recordio_reader_creator("/pfs/dlnel/public/dataset/uci_housing/uci_housing_train*"), buf_size=500),
            batch_size=2),
        feeding=feeding,
        event_handler=event_handler,
        num_passes=30)


if __name__ == '__main__':
    main()
