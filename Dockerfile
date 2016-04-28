FROM buildpack-deps:xenial-curl

RUN mkdir -p /etc/sphinx

COPY ./bin/sphinxd /usr/bin/sphinxd

CMD ["sphinxd", "--config", "/etc/sphinx/sphinx.yaml"]
