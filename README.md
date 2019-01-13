# Accelerator-K8S

Accelerator-K8S is a device plugin works with [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)

## Getting Started
#### Dependencies
- Kubernetes
- [Accelerator-Docker](https://github.com/KAIST-NCL/Accelerator-Docker)


#### Installing
Install using YAML --not supported yet

Clone this repository and make it.
```bash
$ git clone https://github.com/KAIST-NCL/Accelerator-K8S.git
$ cd Accelerator-K8S
$ make
$ sudo make install
```

## How to run Accelerator-K8S
It is a simple daemon, so you can just run it like
```bash
$ out/acc-k8s
```
or make it as a daemon set of kubernetes (will be supported soon)

## Authors
#### KAIST NCL
* [Sunghyun Kim](https://github.com/cqbqdd11519)

## License
