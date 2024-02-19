FROM golang:latest AS compiling
RUN mkdir -p /go/src/comments
WORKDIR /go/src/apigateway
ADD . .
WORKDIR /go/src/apigateway/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest

WORKDIR /root/
COPY --from=compiling /go/src/apigateway/cmd/app .
ARG newsAggregator=http://192.168.1.4:8080
ENV newsAggregator="${newsAggregator}"
ARG	commentsService=http://192.168.1.4:9595
ENV commentsService="${commentsService}"
ARG	cersorService=http://192.168.1.4:8787
ENV cersorService="${cersorService}"

CMD ["./app"]
EXPOSE 8080