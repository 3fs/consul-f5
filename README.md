# Auto-scaled load balancing F5

  * Automatic container registration using [Consul](https://hub.docker.com/r/progrium/consul/) and [Registrator](https://hub.docker.com/r/gliderlabs/registrator/)
  * Load balancing containers with [F5-BIG-IP](https://f5.com) and [Consul-templates](https://github.com/hashicorp/consul-template)

## Install vagrant plugins

```bash
$ vagrant plugin install vagrant-docker-compose
```

## Usage

Create a `Vagrantfile` and a shell script which installs [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/). Take a look at `example` directory in this repository.

In Vagrantfile, specify path to script that will be executed on guest machine:

```ruby
config.vm.provision "shell", path: "~/init.sh"
```

Copy `consul_server` and `consul_worker` to `consul_bigip` folder inside your home directory.

Sync `consul_bigip` folder to the guest machine:

```ruby
server.vm.synced_folder "~/consul_bigip" , "/vagrant/consul_template"
```

To run server machine with docker-compose on `vagrant up`:

```ruby
server.vm.provision :docker
server.vm.provision :docker_compose, yml: "/vagrant/consul_template/consul_server/docker-compose.yml", rebuild: true, run: "always"
```

To run worker machine with docker-compose on `vagrant up`:

```ruby
server.vm.provision :docker
server.vm.provision :docker_compose, yml: "/vagrant/consul_template/consul_worker/docker-compose.yml", rebuild: true, run: "always"
```

## Composed worker machine

Worker machine runs Consul for service discovery, health check, and key/value registry for datacenter nodes and registrator for automatic service registration.

```yaml
consul:
    image: "progrium/consul:latest"
    container_name: "consul"
    hostname: "worker-${HOST_IP_E}"
    ports:
      - "${HOST_IP_E}:8300:8300"
      - "${HOST_IP_E}:8301:8301"
      - "${HOST_IP_E}:8301:8301/udp"
      - "${HOST_IP_E}:8302:8302"
      - "${HOST_IP_E}:8302:8302/udp"
      - "${HOST_IP_E}:8400:8400"
      - "${HOST_IP_E}:8500:8500"
      - "${DNS_IP_E}:53:53/udp"
    command: "-server -advertise ${HOST_IP_E} -join ${SERVER_IP}"

registrator:
    image: gliderlabs/registrator:master
    hostname: registrator
    links:
      - consul:consul
    volumes:
      - "/var/run/docker.sock:/tmp/docker.sock"
    command: -internal consul://consul:8500
```

## Composed load-balance machine

Load-balance machine runs Consul, registrator and Consul Template for parsing consul service catalog and generating JSON data.

```yaml
consul_template:
    build: ./consul-template
    image: consul-template
    container_name: consul_template
    hostname: template
    ports:
      - 80:80
    links:
      - consul:consul
    volumes:
      - "./consul-template/config/bigip.json:/tmp/bigip/bigip.json"
      - "./consul-template/config/bigip.ctmpl:/tmp/bigip/bigip.ctmpl"
    command: "consul-template -consul=consul:8500 -config=/tmp/bigip/bigip.json"

consul:
    image: "progrium/consul:latest"
    container_name: consul
    hostname: "loadbalancer"
    ports:
      - "${HOST_IP_E}:8300:8300"
      - "${HOST_IP_E}:8301:8301"
      - "${HOST_IP_E}:8301:8301/udp"
      - "${HOST_IP_E}:8302:8302"
      - "${HOST_IP_E}:8302:8302/udp"
      - "${HOST_IP_E}:8400:8400"
      - "${HOST_IP_E}:8500:8500"
      - "${DNS_IP_E}:53:53/udp"
    command: "-server -advertise ${HOST_IP_E} -bootstrap-expect 1"

registrator:
    image: gliderlabs/registrator:master
    container_name: registrator
    hostname: registrator
    links:
      - consul:consul
    volumes:
      - "/var/run/docker.sock:/tmp/docker.sock"
    command: -internal consul://consul:8500
```

## How it works

When new node is registered to Consul or new service is added to Consul catalog, Consul Template will generate JSON data from Consul catalog and F5-controller will be called. F5-controller will generate REST calls and send them to F5 BIG IP. F5 BIG IP will be auto configured based on Consul catalog.
