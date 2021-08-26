# (C) Copyright IBM Corp. 2021
#
# SPDX-License-Identifier: Apache-2.0

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4

ENV HOME=/mgmt-api-release/

# shadow-utils.x86_64 installs groupadd
RUN microdnf update -y && \
    microdnf -y install shadow-utils.x86_64 && \
    microdnf clean all

COPY src/hri ${HOME}

WORKDIR mgmt-api-release

# Setup hri user
RUN groupadd -g 1000 hri && \
    useradd --shell /bin/bash -u 1000 -g 1000 -m hri && \
    chown -R hri:hri /mgmt-api-release

USER hri

EXPOSE 1323

ENTRYPOINT ["./hri"]
