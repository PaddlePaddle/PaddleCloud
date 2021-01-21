#!/usr/local/bin/python3
import time, subprocess
from kubernetes import client, config


config.load_incluster_config()
ips = []
while len(ips) < 2:
    time.sleep(10)
    for i in client.CoreV1Api().list_pod_for_all_namespaces(watch=False).items:
        if i.metadata.namespace == "default":
            ips.append(i.status.pod_ip)

ips.sort()
server_ips = ips[0] + ":7164"
worker_ips = ",".join(ips[1:])
subprocess.run(["fleetrun",
                "--servers={}".format(server_ips),
                "--workers={}".format(worker_ips),
                "train.py"])
