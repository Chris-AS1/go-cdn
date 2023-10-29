FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-cdn

FROM golang:1.21-alpine as runtime
WORKDIR /app

COPY --from=builder /app/go-cdn /go-cdn

EXPOSE 3000

ENTRYPOINT [ "/go-cdn" ]
