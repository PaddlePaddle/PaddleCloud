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
    mod = __import__("paddle.v2.dataset." + name, fromlist=[''])

    path = os.path.join(output_path, name)
    mkdir_p(path)

    mod.convert(path)

if __name__ == '__main__':
    a = ['uci_housing']
    for m in a:
        convert("./data", m)
