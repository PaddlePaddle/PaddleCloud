#!/bin/env python
import os
import sys
import time
import socket
from kubernetes import client, config
NAMESPACE = os.getenv("NAMESPACE")
if os.getenv("KUBERNETES_SERVICE_HOST", None):
    config.load_incluster_config()
else:
    config.load_kube_config()
v1 = client.CoreV1Api()


def fetch_pods_info(label_selector, phase=None):
    api_response = v1.list_namespaced_pod(
        namespace=NAMESPACE, pretty=True, label_selector=label_selector)
    pod_list = []
    for item in api_response.items:
        if phase is not None and item.status.phase != phase:
            continue
        pod_list.append((item.status.phase, item.status.pod_ip))
    return pod_list


def wait_pods_running(label_selector, desired):
    print "label selector: %s, desired: %s" % (label_selector, desired)
    while True:
        count = count_pods_by_phase(label_selector, 'Running')
        # NOTE: pods may be scaled.
        if count >= int(desired):
            break
        print 'current cnt: %d sleep for 5 seconds...' % count
        time.sleep(5)


def count_pods_by_phase(label_selector, phase):
    pod_list = fetch_pods_info(label_selector, phase)
    return len(pod_list)


def fetch_ips(label_selector, port=None, phase=None):
    pod_list = fetch_pods_info(label_selector, phase)
    ips = [item[1] for item in pod_list]
    ips.sort()

    r=[]
    for ip in ips:
        if port != "0":
            ip = "{0}:{1}".format(ip, port)

        r.append(ip)

    return ",".join(r)

def fetch_id(label_selector, phase=None):
    pod_list = fetch_pods_info(label_selector, phase)
    ips = [item[1] for item in pod_list]
    ips.sort()
    local_ip = socket.gethostbyname(socket.gethostname())
    for i in xrange(len(ips)):
        if ips[i] == local_ip:
            return i
    return None

def fetch_trainer_ips(label_selector, port=None, phase=None):
    return fetch_ips(label_selector, port, phase)

def fetch_pserver_ips(label_selector, port=None, phase=None):
    return fetch_ips(label_selector, port, phase)

def fetch_trainer_id(label_selector):
    return fetch_id(label_selector, phase="Running")

def fetch_pserver_id(label_selector):
    return fetch_id(label_selector, phase="Running")

def fetch_master_ip(label_selector):
    pod_list = fetch_pods_info(label_selector, phase="Running")
    master_ips = [item[1] for item in pod_list]
    return master_ips[0]


if __name__ == "__main__":
    command = sys.argv[1]
    if command == "fetch_pserver_ips":
        print fetch_pserver_ips(sys.argv[2], sys.argv[3], sys.argv[4])
    if command == "fetch_trainer_ips":
        print fetch_trainer_ips(sys.argv[2], sys.argv[3], sys.argv[4])
    elif command == "fetch_pserver_id":
        print fetch_pserver_id(sys.argv[2])
    elif command == "fetch_trainer_id":
        print fetch_trainer_id(sys.argv[2])
    elif command == "fetch_master_ip":
        print fetch_master_ip(sys.argv[2])
    elif command == "count_pods_by_phase":
        print count_pods_by_phase(sys.argv[2], sys.argv[3])
    elif command == "wait_pods_running":
        wait_pods_running(sys.argv[2], sys.argv[3])
