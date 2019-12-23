# DaBloog - Go Based Blogging Software

Sample blogging software for testing application deployments

# Instructions

```bash
CACHE_URI="host.docker.internal:6379" DB_URI="mongodb://host.docker.internal:27017" go run blog.go
```

## Installation, Instructions

```bash
go build -o blog blog.go
```

### Docker

```
docker build .
```


