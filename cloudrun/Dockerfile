FROM golang:1.21rc2-bullseye AS builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app .


FROM gcr.io/distroless/base-debian11

WORKDIR /root
COPY --from=builder /workspace/app . 

CMD ["./app"]
