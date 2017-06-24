import uci_housing
import pcloud
import importlib
def fetch_all():
    for module_name in filter(lambda x: not x.startswith("__"),
                              dir(pcloud.dataset)):
        if "fetch" in dir(
                importlib.import_module("pcloud.dataset.%s" % module_name)):
            getattr(
                importlib.import_module("pcloud.dataset.%s" % module_name),
                "fetch")()
