FROM golang:1.22-alpine AS dev

RUN go install github.com/air-verse/air@v1.52.3
WORKDIR /app
CMD ["air"]

FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/pulseops ./cmd/server

FROM alpine:3.19 AS prod

WORKDIR /app
COPY --from=build /out/pulseops /app/pulseops
EXPOSE 8080
CMD ["/app/pulseops"]
