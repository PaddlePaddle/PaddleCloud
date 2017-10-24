import sys
import os
import errno
import recordio
import paddle.v2.dataset as ds

def convert(output_path, name):
    mod = __import__("paddle.v2.dataset." + name, fromlist=[''])

    path = os.path.join(output_path, name)

    mod.convert(path)

if __name__ == '__main__':
    a = ['uci_housing']
    for m in a:
        convert("./data", m)
