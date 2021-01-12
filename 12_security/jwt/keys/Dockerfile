FROM alpine:3

COPY ./generate.sh /

RUN set -eux; \
    apk add --no-cache openssl; \
    mkdir /keys; \
    adduser -D generator; \
    chown generator:generator /keys; \
    chown generator:generator /generate.sh; \
    chmod +x /generate.sh

USER generator
WORKDIR /keys

CMD ["/generate.sh"]