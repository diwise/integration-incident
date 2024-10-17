FROM --platform=$BUILDPLATFORM docker.io/golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

WORKDIR /app/cmd/integration-incident

ARG TARGETARCH
ARG TARGETOS
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build

FROM registry.access.redhat.com/ubi9/ubi-minimal

WORKDIR /opt/diwise

COPY --from=builder /app/cmd/integration-incident/integration-incident /opt/diwise/

RUN chmod 775 /opt/diwise

EXPOSE 8080

USER 1001

ENTRYPOINT ["/opt/diwise/integration-incident"]