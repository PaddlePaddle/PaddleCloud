import paddle.v2 as paddle
import os
import gzip
from paddle.v2.reader.creator import cloud_reader
import paddle.v2.dataset.uci_housing as uci_housing

etcd_ip = os.getenv("ETCD_IP")
etcd_endpoint = "http://" + etcd_ip + ":" + "2379"
trainer_id = int(os.getenv("PADDLE_INIT_TRAINER_ID"))

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
        cost=cost, 
        parameters=parameters, 
        update_equation=optimizer, 
        is_local=False, 
        pserver_spec=etcd_endpoint,
        use_etcd=True)

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
            paddle.reader.shuffle(cloud_reader(
                ["/pfs/dlnel/public/dataset/uci_housing/uci_housing_train-*"],
                etcd_endpoint), buf_size=500),
            batch_size=2),
        feeding=feeding,
        event_handler=event_handler,
        num_passes=30)


if __name__ == '__main__':
    main()