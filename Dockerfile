# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
LABEL org.opencontainers.image.source=https://github.com/arnabmitra/tasty-links
WORKDIR /src

# Download dependencies as a separate step for caching
COPY go.mod ./
RUN go mod download -x

ARG TARGETARCH

# Copy source code and build statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /bin/server main.go

################################################################################
# Runtime stage
FROM alpine:latest AS final

LABEL org.opencontainers.image.source=https://github.com/arnabmitra/tasty-links

# Install TLS certificates and timezone data (vital for HTTPS proxying to tastytrade API)
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        mailcap \
        && \
        update-ca-certificates

# Run as non-privileged app user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

WORKDIR /app

# Copy binary and web assets
COPY --from=build /bin/server /bin/server
COPY ./example_dashboard.html ./example_dashboard.html
COPY ./templates ./templates
COPY ./static ./static

USER appuser

EXPOSE 8080

ENTRYPOINT [ "/bin/server" ]
