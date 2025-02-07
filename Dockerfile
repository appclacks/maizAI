FROM golang:1.24.1 AS build-env

ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

# -----------------------------------------------------------------------------

FROM alpine:3.21.3

RUN addgroup -S -g 10000 maizai \
 && adduser -S -D -u 10000 -s /sbin/nologin -G maizai maizai

RUN mkdir /app
RUN chown -R 10000:10000 /app

USER 10000

COPY --from=build-env /app/maizai /app/maizai

ENTRYPOINT ["/app/maizai"]
CMD ["server"]
