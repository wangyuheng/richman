FROM golang:alpine as builder

ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE=on

WORKDIR /go/apps
COPY . /go/apps

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .

FROM alpine AS runner

WORKDIR /go/richman
COPY --from=builder /go/apps/richman .

EXPOSE 80
ENTRYPOINT ["./richman"]