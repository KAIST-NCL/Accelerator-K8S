# Accelerator-K8S

Accelerator-K8S is a device plugin works with [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)

## Getting Started
#### Dependencies
- Kubernetes
- [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)


#### Installing
Deploy as a daemonset **(Recommended)**
```bash
kubectl create -f https://raw.githubusercontent.com/KAIST-NCL/Accelerator-K8S/master/acc-k8s.yml
```

OR

Install manually by cloning and building
```bash
$ git clone https://github.com/KAIST-NCL/Accelerator-K8S.git
$ cd Accelerator-K8S
$ make
$ sudo make install

$ out/acc-k8s
```
for every node

## Authors
#### KAIST NCL
* [Sunghyun Kim](https://github.com/cqbqdd11519)

## License
