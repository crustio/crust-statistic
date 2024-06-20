# Statistic

# Installation

## Building

`make build`: Builds `statistic` in `./build`.

**or**

`make install`: Uses `go install` to add `statistic` to your GOBIN.

## Docker
The official Statistic Docker image can be found here.

To build the Docker image locally run:

```
make build-docker
```

To start ChainBridge:

``` 
docker run -v ./config.ini:/app/config.ini crustio/statistic
```
