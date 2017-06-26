import sys
import os
import errno
import recordio
import paddle.v2.dataset as ds

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:  # Python >2.5
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def convert(output_path, name):
    print "proc " + name
    mod = __import__("paddle.v2.dataset." + name, fromlist=[''])

    path = output_path + "/" + name
    mkdir_p(path)

    if  name == 'mq2007' or name == "sentiment":
        ds.common.convert(path, mod.train, 10, name + "_train")
        ds.common.convert(path, mod.test, 10, name + "_test")
    elif name == 'conll05':
        ds.common.convert(path, mod.test(), 10, name + "_train")
        ds.common.convert(path, mod.test(), 10, name + "_test")
    elif name == 'imdb':
        word_dict = ds.imdb.word_dict()
        ds.common.convert(path, lambda:mod.train(word_dict), 10, name + "_train")
        ds.common.convert(path, lambda:mod.test(word_dict), 10, name + "_test")
    elif name == 'imikolov':
        N=5
        word_dict = ds.imikolov.build_dict()
        ds.common.convert(path, mod.train(word_dict, N), 10, name + "_train")
        ds.common.convert(path, mod.test(word_dict, N), 10, name + "_test")
    elif name == 'cifar':
        ds.common.convert(path, mod.train100(), 10, name + "_train100")
        ds.common.convert(path, mod.test100(), 10, name + "_test100")
        ds.common.convert(path, mod.train10(), 10, name + "_train10")
        ds.common.convert(path, mod.test10(), 10, name + "_test10")
    elif name == 'wmt14':
        dict_size = 30000
        ds.common.convert(path, mod.train(dict_size), 10, name + "_train")
        ds.common.convert(path, mod.test(dict_size), 10, name + "_test")
    else:
        ds.common.convert(path, mod.train(), 10, name + "_train")
        ds.common.convert(path, mod.test(), 10, name + "_test")


if __name__ == '__main__':
    if len(sys.argv) != 2:
        sys.exit("input format:python convert.py output_path")

    output_path=sys.argv[1]
    print "output path:" + output_path 

    a = ['cifar', 'conll05', 'imdb', 'imikolov', 'mnist', 'movielens', 'sentiment', 'uci_housing', 'wmt14']
    for m in a:
        convert(output_path, m)
