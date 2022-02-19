FROM golang as builder
WORKDIR /build
COPY . .
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0
RUN make build

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /build/goosball /app/goosball
COPY static /app/static
EXPOSE 8080
ENTRYPOINT [ "/app/goosball" ]
