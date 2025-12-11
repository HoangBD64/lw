FROM --platform=linux/amd64 golang:1.24-alpine AS gobuilder
WORKDIR /backend
ENV CGO_ENABLED=0
COPY . .
RUN go mod tidy
RUN mkdir -p /out/

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -buildvcs=false -o /out/server ./cmd/server

# FROM scratch AS bin
FROM alpine:3.15 AS bin
WORKDIR /app

COPY --from=gobuilder ./backend/config.yaml .
COPY --from=gobuilder /out/server server

CMD ["./server"]
