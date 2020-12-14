FROM golang:1.14

WORKDIR /go/src/app
COPY . .
ARG db_url
ENV env_var_name=$db_url

RUN go get -d -v ./...
RUN go build -v ./...

CMD ["./api"]
