FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/givememoney.fun-backend ./cmd/main.go

FROM alpine:3.18
COPY --from=builder /bin/givememoney.fun-backend  /bin/givememoney.fun-backend

EXPOSE 9090
ENTRYPOINT [ "/bin/givememoney.fun-backend"]
