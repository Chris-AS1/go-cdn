FROM golang:1.18-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./

# Will be mounted
RUN mkdir ./resources

RUN go mod download

COPY *.go ./

RUN go build -o /go-cdn

EXPOSE 3333
CMD [ "/go-cdn" ]