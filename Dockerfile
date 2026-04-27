FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod ./
COPY main.go ./
RUN go build -o /bin/data-benchmark-golang .

FROM alpine:3.22
WORKDIR /app
COPY --from=build /bin/data-benchmark-golang /usr/local/bin/data-benchmark-golang
EXPOSE 8080
CMD ["data-benchmark-golang"]
