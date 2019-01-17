FROM grpc/go:1.0

RUN apt-get update && \
     apt-get install -y --no-install-recommends make&& \
     rm -rf /var/lib/apt/lists/* && \
     git clone https://github.com/KAIST-NCL/Accelerator-K8S.git /ACC-K8S && \
     cd /ACC-K8S && \
     make && \
     cp /ACC-K8S/out/acc-k8s / && \
     rm -rf /ACC-K8S && \
     apt-get purge -y make

ENTRYPOINT ["/acc-k8s"]