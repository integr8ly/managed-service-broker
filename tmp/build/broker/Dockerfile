FROM registry.access.redhat.com/ubi8/ubi:8.4

RUN dnf install ca-certificates

RUN adduser managed-service && passwd -d managed-service
USER managed-service

ADD tmp/_output/bin/managed-service-broker /usr/local/bin/

ENTRYPOINT ["managed-service-broker"]