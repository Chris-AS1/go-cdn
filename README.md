# Golang CDN
Simple CDN made in Golang which aims to distribute files from a simple and easily deployable environment.

# Parameters
| VAR | Description |
|---|---|
| `CDN_PORT=3333` | Port used by CDN |  
| `CDN_SUBPATH=/image/` | Subpath on which the resources will be served |  
| `CDN_SUBPATH_ENABLE=true` | Select if CDN should serve under a subpath, NOT RECOMMENDED |  
| `CDN_ENABLE_DELETE=true` | Enables the DELETE endpoint on all the images |  
| `CDN_ENABLE_INSERTION=false` | Disables the POST endpoints |  
| `DB_USERNAME=`| Auth Username for PostgreSQL |  
| `DB_PASSWORD=` | Auth Password for PostgreSQL |  
| `DB_PORT=` | Port for PostgreSQL |  
| `DB_URL=` | Database URL for PostgreSQL |  
| `DB_NAME=` | Database Name for PostgreSQL |  
| `DB_COL_ID=` | Table Column containing IDs |  
| `DB_COL_FN=` | Table Column containing Image File Name |  
| `DB_SSL=enabled\|disabled` | SSL Options for PostgreSQL |  

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
```docker
version: '3.3'
services:
    go-cnd:
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
- [ ] Image Fixed Hash as ID - on Redis
- [ ] Caching Redis
- [x] Option to disable subpath
- [x] File Mapping with ID
- [ ] Auth
- [ ] Compresssion
- [ ] Geo Restriction
- [ ] Postgres for Authentication

