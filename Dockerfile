FROM golang:alpine AS binary
RUN apk add --no-cache libc-dev gcc git sqlite-dev
WORKDIR /go/src
RUN git clone https://github.com/pNre/FeedSpammer feedspammer
WORKDIR /go/src/feedspammer
RUN go get
RUN go build -o main

FROM alpine:3.6
RUN apk add --no-cache ca-certificates && update-ca-certificates
WORKDIR /app
ENV PORT 8000
EXPOSE 8000
COPY --from=binary /go/src/feedspammer/main /app
VOLUME ["/database"]
CMD ["/app/main", "/database/main.db"]
