FROM alpine:3.6

RUN apk add --no-cache ca-certificates libc6-compat

RUN adduser -D managed-service
USER managed-service

ADD tmp/_output/bin/managed-service-broker /usr/local/bin/

ENTRYPOINT ["managed-service-broker"]