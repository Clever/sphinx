FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates

RUN mkdir -p /etc/sphinx

COPY ./bin/sphinxd /usr/bin/sphinxd

COPY kvconfig.yml /kvconfig.yml

CMD ["sphinxd", "--config", "/etc/sphinx/sphinx.yaml"]
