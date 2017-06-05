import paddle.v2.dataset as dataset
import shutil
import os
dataset_home = os.getenv("DATASET_HOME")
os.system("mv %s %s" % (dataset.common.DATA_HOME, dataset_home))

dataset.common.DATA_HOME = dataset_home
dataset.common.split(dataset.uci_housing.train(),
                    line_count = 500,
                    suffix=dataset_home + "/uci_housing/train-%05d.pickle")
dataset.common.split(dataset.mnist.train(),
                    line_count = 500,
                    suffix=dataset_home + "/mnist/train-%05d.pickle")
dataset.common.split(dataset.cifar.train10(),
                    line_count = 500,
                    suffix=dataset_home + "/cifar/train10-%05d.pickel")

N = 5
word_dict = dataset.imikolov.build_dict()
dict_size = len(word_dict)
dataset.common.split(dataset.imikolov.train(word_dict, 5),
                    line_count = 500,
                    suffix = dataset_home + "/imikolov/train-%05d.pickle")

dataset.common.split(dataset.movielens.train(),
                    line_count = 500,
                    suffix = dataset_home + "/movielens/train-%05d.pickle")

dataset.common.split(lambda: dataset.imdb.train(dataset.imdb.word_dict()),
                    line_count = 500,
                    suffix = dataset_home + "/imdb/train-%05d.pickle")

dataset.common.split(dataset.conll05.test(),
                    line_count = 500,
                    suffix = dataset_home + "/conll05/test-%05d.pickle")

dataset.common.split(dataset.wmt14.train(30000),
                    line_count = 500,
                    suffix = dataset_home + "wmt14/train-%05d.pickle")
