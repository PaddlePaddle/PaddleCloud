import paddle.v2 as paddle
import paddle.v2.dataset as dataset
import os
import gzip
import sys

# NOTE: You should full fill your username, for example:
#   USERNAME = "paddle@example.com"
# TODO(Yancey1989): fetch username from environment variable.
USERNAME = "YOUR USERNAME"

DC = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")

#PaddleCloud cached the dataset on /pfs/${DATACENTER}/home/${USERNAME}/...
dataset.common.DATA_HOME = "/pfs/%s/home/%s" % (DC, USERNAME)
TRAIN_FILES_PATH = os.path.join(dataset.common.DATA_HOME, "uci_housing")

def prepare_dataset():
    dataset.common.convert(TRAIN_FILES_PATH,
                           dataset.uci_housing.train(), 100, "train")

TRAINER_ID = int(os.getenv("PADDLE_INIT_TRAINER_ID"))
TRAINER_INSTANCES = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS"))

def cluster_reader_recordio(paths, trainer_id, trainer_instances):
    """
    Creates a cluster data reader from given RecordIO file paths,
        each trainer will read a subset of the whole files set.

    :paths: path of recordio files.
    :trainer_id: current trainer ID.
    :trainer_instances: total trainer instances count.
    :returns data reader of RecordIO files.
    """

    import recordio as rec
    import pickle
    import glob

    def reader():
        file_list = glob.glob(paths)
        file_list.sort()
        my_file_list = []
        # collect a subset files according with the trainer_id
        for idx, f in enumerate(file_list):
            if idx % trainer_instances == trainer_id:
                my_file_list.append(f)
        for f in my_file_list:
            print "processing", f
            reader = rec.reader(f)
            record_raw = reader.read()
            while record_raw:
                yield pickle.loads(record_raw)
                record_raw = reader.read()
            reader.close()
    return reader

def main():
    # init
    paddle.init()

    # network config
    x = paddle.layer.data(name='x', type=paddle.data_type.dense_vector(13))
    y_predict = paddle.layer.fc(input=x, size=1, act=paddle.activation.Linear())
    y = paddle.layer.data(name='y', type=paddle.data_type.dense_vector(1))
    cost = paddle.layer.square_error_cost(input=y_predict, label=y)

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
                reader=paddle.batch(dataset.uci_housing.test(), batch_size=2),
                feeding=feeding)
            print "Test %d, Cost %f" % (event.pass_id, result.cost)
            if TRAINER_ID == "0":
                with gzip.open("fit-a-line_pass_%05d.tar.gz" % event.pass_id,
                               "w") as f:
                    parameters.to_tar(f)
    # training
    trainer.train(
        reader=paddle.batch(
            paddle.reader.shuffle(
                cluster_reader_recordio(
                    os.path.join(TRAIN_FILES_PATH, "train-*"),
                    TRAINER_ID,
                    TRAINER_INSTANCES),
                buf_size=500),
            batch_size=2),
        feeding=feeding,
        event_handler=event_handler,
        num_passes=30)


if __name__ == '__main__':
    usage = "python train.py [prepare|train]"
    if len(sys.argv) != 2:
        print usage
        exit(1)

    if TRAINER_ID == -1 or TRAINER_INSTANCES == -1:
        print "no cloud environ found, must run on cloud"
        exit(1)

    if sys.argv[1] == "prepare":
        prepare_dataset()
    elif sys.argv[1] == "train":
        main()
