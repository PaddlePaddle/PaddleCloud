import uci_housing
import paddle
import importlib
def fetch_all():
    for module_name in filter(lambda x: not x.startswith("__"),
                              dir(paddle.cloud.dataset)):
        if "fetch" in dir(
                importlib.import_module("paddle.cloud.dataset.%s" % module_name)):
            getattr(
                importlib.import_module("paddle.cloud.dataset.%s" % module_name),
                "fetch")()
