FROM golang:1.22.4 as builder

ARG PROMU_VERSION=0.17.0

WORKDIR /src

RUN curl -L -o /tmp/promu-${PROMU_VERSION}.linux-amd64.tar.gz https://github.com/prometheus/promu/releases/download/v${PROMU_VERSION}/promu-${PROMU_VERSION}.linux-amd64.tar.gz && \
    tar -xvf /tmp/promu-${PROMU_VERSION}.linux-amd64.tar.gz -C /tmp && \
    mv /tmp/promu-${PROMU_VERSION}.linux-amd64/promu /usr/bin/promu

COPY . /src

RUN /usr/bin/promu build && ls /src

FROM gcr.io/distroless/static

COPY --from=builder /src/openstack_exporter /bin/openstack_exporter

ENTRYPOINT ["/bin/openstack_exporter"]
