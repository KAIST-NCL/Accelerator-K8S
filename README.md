# Accelerator-K8S

Accelerator-K8S is a device plugin works with [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)

## Getting Started
### Dependencies
- Kubernetes
- [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)


### Installation
- Deploy as a daemonset **(Recommended)**
```bash
kubectl create -f https://raw.githubusercontent.com/KAIST-NCL/Accelerator-K8S/master/acc-k8s.yml
```

OR

- Install manually by cloning and building
```bash
$ git clone https://github.com/KAIST-NCL/Accelerator-K8S.git
$ cd Accelerator-K8S
$ make
$ sudo make install

$ out/acc-k8s
```
for every node

## How to Use
### Prerequisites
First, you need to make sure that you installed [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker) properly.

* `default-runtime` of docker daemon should be set to `acc-runtime`. Check `/etc/docker/daemon.json`
* In Accelerator-Docker setting(/etc/accelerator-docker/device.pbtxt), **accelerator type** should be named following
kubernetes resourece naming rule. Refer to [this](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/).

### Enabling Accelerators in Containers
You can require your accelerators as a resource like below:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-pod
spec:
  containers:
    - name: demo-container
      image: ubuntu:16.04
      resources:
        limits:
          {ACCELERATOR_TYPE}: 1
```
'Accelerator type' is configured in Accelerator-Docker configuration file and you can check it by using `acc-manager list` command

* Example) 
```bash
$ acc-manager list
+------------+--------------+------------------------+----------------+----------------+----------+
|     ID     |     Name     |          Type          |    PCI-Slot    |     Status     |  Holder  |
+------------+--------------+------------------------+----------------+----------------+----------+
| 234750     | QuadroM2000  | nvidia.com/gpu         | 0000:02:00.0   | Available      | 0        |
| 1121730    | KCU-1500     | xilinx.fpga/kcu-1500   | 0000:01:00.0   | Available      | 0        |
+------------+--------------+------------------------+----------------+----------------+----------+
```
If you want to use KCU-1500 for your pod, you can set container limit as follows.
```yaml
  containers:
    - name: demo-container
      image: ubuntu:16.04
      resources:
        limits:
          xilinx.fpga/kcu-1500: 1
```

## Authors
#### KAIST NCL
* [Sunghyun Kim](https://github.com/cqbqdd11519)

## License
This project is released under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0), see [LICENSE](LICENSE) for more information.