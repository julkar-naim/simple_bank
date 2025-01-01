# Build stage
FROM golang:1.23-alpine AS base
RUN apk add --no-cache curl tar
WORKDIR /app
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.11.0/migrate.linux-amd64.tar.gz | tar xvz -C /tmp && \
    mv /tmp/migrate.linux-amd64 /bin/migrate

FROM base AS development
# install go air
RUN curl -L https://github.com/air-verse/air/releases/download/v1.61.5/air_1.61.5_linux_amd64 --output /tmp/air && \
    mv /tmp/air /bin/air && \
    chmod +x /bin/air
EXPOSE 8080

# build stage
FROM base AS builder
COPY . .
RUN go mod download
RUN go build -o main main.go

# production stage
FROM alpine:3.21
WORKDIR /app
COPY --from=base /bin/migrate /bin/migrate
COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY db/migration ./db/migration

EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ]