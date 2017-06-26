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

    mod.convert(path)

if __name__ == '__main__':
    if len(sys.argv) != 2:
        sys.exit("input format:python convert.py output_path")

    output_path=sys.argv[1]
    print "output path:" + output_path 

    a = ['cifar', 'conll05', 'imdb', 'imikolov', 'mnist', 'movielens', 'sentiment', 'uci_housing', 'wmt14']
    for m in a:
        convert(output_path, m)
