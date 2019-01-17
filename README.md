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
          acc.k8s/{DEVICE_NAME}: 1
```
All alphabets of device name should be lower case, even if it is not set as lower case in Accelerator-Docker configuration.

## Authors
#### KAIST NCL
* [Sunghyun Kim](https://github.com/cqbqdd11519)

## License
