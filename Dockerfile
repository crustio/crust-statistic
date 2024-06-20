FROM  golang:1.18-stretch AS builder
ADD . /src
WORKDIR /src
RUN go mod download
RUN go build  .

FROM debian:stretch-slim
RUN apt-get -y update && apt-get -y upgrade && apt-get install ca-certificates wget -y


COPY --from=builder /src/statistic /app/statistic
COPY --from=builder /src/config.ini /app/config.ini

RUN chmod +x /app/statistic
WORKDIR /app
ENTRYPOINT ["./statistic"]
