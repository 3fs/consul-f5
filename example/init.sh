#!/bin/bash
echo "Updating apt-get"
apt-get update -qq
echo "installing https"
apt-get install -qq apt-transport-https ca-certificates
echo "inserting key"
apt-key adv -qq --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D

echo "deb https://apt.dockerproject.org/repo ubuntu-precise main" > /etc/apt/sources.list.d/docker.list
apt-get update -qq
echo "installing latest docker engine "
apt-get purge -qqy lxc-docker

apt-get install -qqy docker-engine

apt-get upgrade -qqy docker-engine
echo "installing kernel 3.13"
aptitude -y install linux-firmware linux-generic-lts-trusty linux-headers-3.13.0-32 linux-headers-3.13.0-32-generic linux-headers-generic-lts-trusty linux-image-3.13.0-32-generic linux-image-generic-lts-trusty
echo "installing curl"
apt-get install -qqy curl

curl -L https://github.com/docker/compose/releases/download/1.7.0/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
echo "insytalling docker-compose"
chmod +x /usr/local/bin/docker-compose