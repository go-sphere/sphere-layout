FROM golang:1.24.4 AS builder
WORKDIR /app

ARG BUILD_VERSION
ENV BUILD_VERSION=${BUILD_VERSION}
#ENV GOPROXY=https://goproxy.io

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make init && make build

FROM gcr.io/distroless/static
COPY --from=builder /app/build/current_arch/app /main
ENV GIN_MODE=release
ENTRYPOINT ["/main"]
CMD ["--config", "/config/config.json"]