FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN go build -ldflags="-s -w" -o /cron-parser .

FROM alpine:3.20
COPY --from=build /cron-parser /usr/local/bin/cron-parser
ENTRYPOINT ["cron-parser"]
