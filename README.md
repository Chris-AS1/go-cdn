# Golang CDN
Simple CDN made in Golang which aims to distribute files from a simple and easily deployable environment.

# Parameters
| VAR | Description |
|---|---|
| `CDN_PORT=3333` | Port used by CDN |  
| `CDN_SUBPATH=/image/` | Subpath on which the resources will be served |  
| `CDN_SUBPATH_ENABLE=true\|false` | Select if CDN should serve under a subpath, NOT RECOMMENDED |  
| `CDN_ENABLE_DELETE=true\|false` | Select if the DELETE endpoints are enabled |  
| `CDN_ENABLE_INSERTION=true\|false` | Select if the POST endpoints are enabled |  
| `DB_USERNAME=`| Auth Username for PostgreSQL |  
| `DB_PASSWORD=` | Auth Password for PostgreSQL |  
| `DB_PORT=` | Port for PostgreSQL |  
| `DB_URL=` | Database URL for PostgreSQL |  
| `DB_NAME=` | Database Name for PostgreSQL |  
| `DB_COL_ID=` | Table Column containing IDs |  
| `DB_COL_FN=` | Table Column containing Image File Name |  
| `DB_SSL=enable\|disable` | SSL Options for PostgreSQL |  
| `REDIS_URL=redis:6379` | Redis Connection URL (IP:Port) |  


# Docker Deployment
## Building the image
```bash
docker build -t golang/cdn .
```

## Running it
```bash
docker run -p 8080:3333 -v "$(pwd)/resources":/config/resources:ro -e CDN_SUBPATH=/v1/ golang/cdn:latest
```
This will run the CDN with the following specifics:
- Accessible at http://IP:8080/v1/image
- Local `resources` folder mapped to the internal directory in Read Only (Note that it shall be changed for the DELETE endpoint to work). You should always map the folder containing images to `/config/resources` on the container's path


## Compose
Alternatively, using the following `docker-compose.yml` file:
```docker
version: '3.3'
services:
    go-cdn:
        image: 'golang/cdn:latest'
        ports:
            - '8080:3333'
        volumes:
            - PATH/resources:/config/resources:ro
        environment:
            - CDN_SUBPATH=/v1/
```


---

## Todo
- [x] Insert, Remove Images
- [x] Image Fixed Hash as ID - on Redis
- [ ] Caching Redis
- [x] Option to disable subpath
- [x] File Mapping with ID
- [ ] Authentication
- [ ] Compression
- [ ] Geo Restriction
- [ ] Try out Couchbase

