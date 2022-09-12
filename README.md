# Golang CDN
Simple CDN made in Golang which aims to distribute files from a simple environment.

# Parameters
- CDN_PATH
- DB_USERNAME
- DB_PASSWORD
- DB_PORT
- DB_URL
- DB_NAME
- DB_SSL


## Todo
- [ ] Database Storage
- [ ] Caching
- [ ] Auth
- [ ] Geo Restriction

| VAR | Description |
|---|---|
| `CDN_PATH=/image/` | Subpath on which the resources will be served |  
| `DB_USERNAME` | Auth Username for PostgreSQL |  
| `DB_PASSWORD` | Auth Password for PostgreSQL |  
| `DB_PORT` | Port for PostgreSQL |  
| `DB_URL` | Database URL for PostgreSQL |  
| `DB_NAME` | Table Name for PostgreSQL |  
| `DB_SSL=enabled\|disabled` | SSL Options for PostgreSQL |  