FROM golang:1.18-alpine as builder

WORKDIR /config
COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY . .

RUN go build -o /go-cdn

FROM golang:1.21-alpine as runtime
WORKDIR /app

# Will be externally mounted and served
RUN mkdir ./resources

COPY --from=builder /go-cdn /go-cdn

EXPOSE 3333

ENTRYPOINT [ "/go-cdn" ]
