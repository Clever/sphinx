#!/usr/bin/env bash

apt-get update
apt-get install -y make git mercurial
wget https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz -O /tmp/godeb-amd64.tar.gz
tar xvzf /tmp/godeb-amd64.tar.gz
mv godeb /usr/local/bin
godeb install 1.2
mkdir -p /home/vagrant/.go
chown vagrant:vagrant /home/vagrant/.go
echo "export GOPATH=~/.go" >> /home/vagrant/.bashrc
# Git repo is mounted to /vagrant, start there
echo "cd /vagrant" >> /home/vagrant/.bashrc
echo "export PATH=\$PATH:\$GOPATH/bin" >> /home/vagrant/.bashrc
