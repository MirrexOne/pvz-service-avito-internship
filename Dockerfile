FROM golang:alpine AS builder
LABEL stage=builder

ARG MIGRATE_VERSION=v4.17.1
RUN apk add --no-cache curl tar postgresql-client build-base git protobuf protobuf-dev

RUN curl -fL https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz -o migrate.linux-amd64.tar.gz && \
    tar -xzf migrate.linux-amd64.tar.gz && \
    if [ -f "migrate" ]; then mv migrate /usr/local/bin/migrate; \
    elif [ -f "migrate.linux-amd64" ]; then mv migrate.linux-amd64 /usr/local/bin/migrate; \
    else echo "ERROR: Cannot find 'migrate' or 'migrate.linux-amd64' in extracted files" && exit 1; fi && \
    chmod +x /usr/local/bin/migrate && \
    rm migrate.linux-amd64.tar.gz
RUN migrate -version

WORKDIR /app

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
RUN go install github.com/vektra/mockery/v2@latest

ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY scripts ./scripts
RUN chmod +x /app/scripts/generate.sh

RUN /bin/sh /app/scripts/generate.sh

COPY migrations /app/migrations

COPY . .

RUN echo "Generating mocks..." && mockery --all --keeptree && echo "Mocks generated."

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/pvz-server ./cmd/server

FROM alpine:latest
LABEL stage=final

RUN apk add --no-cache ca-certificates libpq netcat-openbsd

WORKDIR /app

COPY --from=builder /app/pvz-server /app/pvz-server
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080 3000 9000
CMD ["/app/pvz-server"]