# Go-CDN
Microservice that serves image BLOBs from PostgreSQL via a REST API, using Redis as cache (LFU), Consul for Service Discovery and HAProxy as Load Balancer (soontm).

## Architecture
![architecture](./assets/architecture_sketch.png)

# Configuration Sample (config.yaml)
```yaml
consul:
  enable: 
  service_name:    # Service under which the microservices will be registered. Each one will have an unique id.
  service_address: # Either auto or manually set. auto finds the first non-loopback address.
  address: 
  datacenter: 
  port: 

redis:
  enable: 
  host:         # If Consul is enabled then this is the service name, otherwise ip:port
  password: 
  db: 

postgres:
  host:         # If Consul is enabled then this is the service name, otherwise ip:port
  database: 
  username:
  password: 
  ssl: 

http:
  allow_insert: 
  allow_delete:
```

# Docker Deployment
## Build the image
```bash
docker build -t local/go-fileserver .
```

## Docker Compose
Check the provided `docker-compose.yml` for a deployment example. The provided stack contains an example Consul container for demo purposes.
```bash
docker compose build 
docker compose up -d
docker compose logs -f
```

# Testing
Due the nature of Go, tests are ran inside their respective packages. This creates confusion with the relative paths regarding configs and migrations.
To get around this limitation it's possible to compile each test individually, and then run it from the root of the folder:
```bash
go test -c ./...
./{PACKAGE}.test 
```
