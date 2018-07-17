FROM alpine:3.6

RUN adduser -D managed-services
USER managed-services

ADD tmp/_output/bin/managed-services-broker /usr/local/bin/

ENTRYPOINT ["managed-services-broker"]