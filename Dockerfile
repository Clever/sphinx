FROM buildpack-deps:xenial-curl

COPY ./bin/sphinxd /usr/bin/sphinxd

CMD ["sphinxd", "--config", "/etc/sphinx/sphinx.yaml"]
