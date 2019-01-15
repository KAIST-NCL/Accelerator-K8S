FROM grpc/go:1.0

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates build-essential

COPY . /ACC-K8S
RUN cd /ACC-K8S && make

ENTRYPOINT ["/ACC-K8S/out/acc-k8s"]