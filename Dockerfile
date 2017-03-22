FROM buildpack-deps:xenial-curl

RUN mkdir -p /etc/sphinx

COPY ./bin/sphinxd /usr/bin/sphinxd

COPY kvconfig.yml /usr/bin/kvconfig.yml

CMD ["sphinxd", "--config", "/etc/sphinx/sphinx.yaml"]
