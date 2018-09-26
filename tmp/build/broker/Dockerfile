FROM alpine:3.6

RUN adduser -D managed-service
USER managed-services

ADD tmp/_output/bin/managed-service-broker /usr/local/bin/

ENTRYPOINT ["managed-service-broker"]