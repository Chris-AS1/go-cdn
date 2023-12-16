FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-cdn ./cmd/go-cdn/main.go 

FROM golang:1.21-alpine as runtime

WORKDIR /cdn

COPY --from=builder /app/go-cdn go-cdn

COPY ./configs/config.yaml.sample /config/config.yaml
COPY ./migrations migrations

EXPOSE 3000

ENTRYPOINT [ "/cdn/go-cdn" ]
