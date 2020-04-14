FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/worker-osp-operator \
    USER_UID=1001 \
    USER_NAME=worker-osp-operator

# install operator binary
COPY build/_output/bin/worker-osp-operator ${OPERATOR}

COPY build/bin /usr/local/bin
COPY bindata /bindata
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
