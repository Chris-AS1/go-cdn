# Golang Fileserver
Microservice implementation of a file-server using PostgreSQL and Redis as databases.
The program uses Consul to handle Service Discovery between instances, Redis as caching mechanism for most requested resources and PostgreSQL as primary Database.

# Configuration Sample (config.yaml)
```yaml
consul:
  reg_service_name: 
  address:
  db: 
  port: 

redis:
  host: 
  port:
  password: 
  enable:
  db:

postgres:
  host:
  port:
  username:
  password:
  ssl:
```


# Docker Deployment
## Building the image
```bash
docker build -t local/go-fs .
```

## Run (TODO)
```bash
docker run -p 8080:3000 -v "$(pwd)/resources":/config/resources:ro -e CDN_SUBPATH=/v1/ golang/cdn:latest
```
This will run the CDN with the following specifics:
- Accessible at http://IP:8080/v1/image
- Local `resources` folder mapped to the internal directory in Read Only (Note that it shall be changed for the DELETE endpoint to work). You should always map the folder containing images to `/config/resources` on the container's path


## Docker Compose
Alternatively, using the `docker-compose.yml` file provided.
Then deploy it with:
```bash
docker compose up -d
```

---

## Todo
- [ ] Insert, Remove Images
- [ ] Image Fixed Hash as ID - on Redis
- [ ] Caching Redis
- [ ] Option to disable subpath
- [x] File Mapping with ID
- [x] Edit redis.configs to implement LRU
- [ ] Authentication
- [ ] Geo Restriction
- [ ] Resize feature via URL parameters
- [ ] Add distributed support
- [ ] Try out Couchbase (?)
