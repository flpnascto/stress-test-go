FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o stresstest ./cmd/cli/main.go

FROM scratch
COPY --from=builder /app/stresstest .
ENTRYPOINT ["./stresstest"]