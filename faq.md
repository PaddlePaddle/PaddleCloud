# FAQ

Before diving in concrete case, some context you may concern,

* the demo *wide_and_deep* runs with cpu only while *resnet* require gpu to run;
* the default install script install paddle-operator under namespace *paddle-system* and watch paddlejob in this ns only;

And please make sure you are using the latest versions of *crd* and *controller* image or at least the consistent versions of the two.

#### Q: Why no paddlejob pod is scheduled ?

Please check the log of controller if the creation was notified, if yes there must be some output as follows,
```
2021-08-03T03:00:56.056Z	INFO	controllers.PaddleJob	Reconcile	{"paddlejob": "paddle-system/resnet", "version": "37646098"}
```
if no, make sure your controller deployed with something like *--namespace=paddle-system* while your paddlejob is apply to other namespace.

#### Q: CreateContainerConfigError ?
After pods creation, it may in CreateContainerConfigError status as,
```
wide-ande-deep-ps-0                        0/1     CreateContainerConfigError   0          3s
wide-ande-deep-worker-0                    0/1     CreateContainerConfigError   0          3s
```
Hard to say it, but this error is shown as expected. 
Since paddle use configmap to exchange global information inter-pods at setup, the creation of configmap depend on the complete creation of ALL pods, 
the pods may shown temporarily in *CreateContainerConfigError* status. This status won't last long in most cases.


#### Q: No information for paddlejob ?

If the output of command shows as follows,
```
kubectl -n paddle-system get pdj
NAME                     STATUS      MODE         AGE
paddle-mnist
```
it may be caused by the incompatible versions of CRD adn controller image, try to update them to current version.

#### Q: Configuration error in log ?
In resnet demo, you may see logs below,

```
-----------  Configuration Arguments -----------
elastic_server: None
force: False
gpus: None
heter_worker_num: None
heter_workers:
host: None
http_port: None
ips: 127.0.0.1
job_id: None
log_dir: log
np: None
nproc_per_node: None
run_mode: None
scale: 0
server_num: None
servers:
training_script: train_fleet_dygraph_ckpt.local.py
training_script_args: []
worker_num: None
workers:
------------------------------------------------
```
Indeed, the information above are correct, it shows the args info in command line, while paddle-operator use environ as configuration.

