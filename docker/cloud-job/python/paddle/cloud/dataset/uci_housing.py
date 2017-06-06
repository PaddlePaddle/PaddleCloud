import paddle.v2.dataset.uci_housing as uci_housing
import paddle.v2.dataset.common as common
import os

__all__=["train", "test", "fetch"]

CURRENT_DATACENTER = os.getenv("CURRENT_DATACENTER")

common.DATA_HOME = "/pfs/%s/public/dataset" % CURRENT_DATACENTER

TRAIN_FILES_PATTERN = os.path.join(common.DATA_HOME,
                                   "uci_housing/train-*.pickle")
TRAIN_FILES_SUFFIX = os.path.join(common.DATA_HOME,
                                  "uci_housing/train-%05d.pickle")


def train():
    return common.cluster_files_reader(
        TRAIN_FILES_PATTERN,
        trainer_count = int(os.getenv("TRAINERS", "1")),
        trainer_id = int(os.getenv("PADDLE_INIT_TRAINER_ID", "0")))

def test():
    return uci_housing.test()

def fetch():
    print "fetch cluster files: %s" % TRAIN_FILES_SUFFIX
    common.split(uci_housing.train(),
                 line_count = 500,
                 suffix=TRAIN_FILES_SUFFIX)
