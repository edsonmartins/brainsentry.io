FROM golang:1.23-alpine AS build
WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o brainsentry ./cmd/server

# Final image
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /app/brainsentry /brainsentry
COPY --from=build /app/config.yaml /config.yaml
COPY --from=build /app/internal/repository/postgres/migrations /migrations

EXPOSE 8080
ENTRYPOINT ["/brainsentry"]
