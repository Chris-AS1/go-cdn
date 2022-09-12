# Golang CDN
Simple CDN made in Golang which aims to distribute files from a simple and easily deployable environment.

# Parameters
| VAR | Description |
|---|---|
| `CDN_PORT=3333` | Port used by CDN |  
| `CDN_PATH=/image/` | Subpath on which the resources will be served |  
| `DB_USERNAME=` | Auth Username for PostgreSQL |  
| `DB_PASSWORD=` | Auth Password for PostgreSQL |  
| `DB_PORT=` | Port for PostgreSQL |  
| `DB_URL=` | Database URL for PostgreSQL |  
| `DB_NAME=` | Table Name for PostgreSQL |  
| `DB_SSL=enabled\|disabled` | SSL Options for PostgreSQL |  

## Todo
- [ ] Select Database Storage (+Redis)
- [ ] Option to disable subpath
- [x] File Mapping with ID
- [ ] Caching
- [ ] Auth
- [ ] Geo Restriction

