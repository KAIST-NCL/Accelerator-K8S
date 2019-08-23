FROM grpc/go:1.0

RUN apt-get update && \
     apt-get install -y --no-install-recommends make&& \
     git clone https://github.com/KAIST-NCL/Accelerator-K8S.git /ACC-K8S && \
     cd /ACC-K8S && \
     make && \
     cp /ACC-K8S/out/acc-k8s / && \
     rm -rf /ACC-K8S && \
     apt-get purge -y make && \
     apt-get clean && \
     apt-get autoclean && \
     apt-get autoremove && \
     rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENTRYPOINT ["/acc-k8s"]