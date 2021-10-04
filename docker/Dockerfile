# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

# Use IBM's standard base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4


# yum is required by the IBM CLI installer and it tries to install git, curl, docker, kubectl, and helm
# The docker install fails and causes problems. We don't need it, so this adds a dummy file to trick
# the installer into thinking that docker is installed.
# jq is used for JSON parsing of Elastic credentials.
RUN microdnf update -y && \
    microdnf -y install git openssl yum which tar sudo && \
    microdnf clean all && \
    echo "echo 'fake docker install'" > /usr/local/bin/docker && \
    chmod 777 /usr/local/bin/docker && \
    curl -sL --fail https://ibm.biz/idt-installer | bash && \
    yum clean all && \
    curl -sL --fail https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 > /usr/local/bin/jq && \
    chmod 755 /usr/local/bin/jq

COPY docker/run.sh \
    docker/elastic.sh \
    docker/appid.sh \
    document-store/index-templates/batches.json \
    deploy.sh \
    mgmt-api-manifest.yml \
    run-smoketests.sh \
    /mgmt-api-release/

COPY build \
    /mgmt-api-release/build/

WORKDIR mgmt-api-release

# Setup hri user 
# Disable priveledge escalation utilities
#   https://github.com/goodwithtech/dockle/blob/master/CHECKPOINT.md#cis-di-0008
RUN groupadd -g 1000 hri && \
    useradd --shell /bin/bash -u 1000 -g 1000 -m hri && \
    chown -R hri:hri /mgmt-api-release && \
    chmod u-s /usr/bin/chage && \
    chmod u-g /usr/bin/chage && \
    chmod u-s /usr/bin/gpasswd && \
    chmod u-g /usr/bin/gpasswd && \
    chmod u-s /usr/bin/newgrp && \
    chmod u-g /usr/bin/newgrp 

USER hri

# IBM CLI plugins have to be installed for each user
RUN ibmcloud plugin install cloud-functions

ENTRYPOINT ["./run.sh"]
