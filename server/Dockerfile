FROM golang:alpine AS builder

WORKDIR /app

COPY . /app

# 移動檔案到 server 目錄
RUN mkdir ./server && mv cmd go.* internal pkg ./server

# 建置二進位檔案
# ENV CGO_ENABLED=1 # For sqlite
# RUN apk add --no-cache gcc musl-dev # For sqlite
RUN cd ./server && go get -u ./... && go mod tidy
RUN cd ./server && GOOS=linux go build -v -o ../main ./cmd


# 最終映像階段
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/server/cmd/.env /app/.env
COPY --from=builder /app/server/cmd/public.pem /app/public.pem

EXPOSE 8080

CMD ["/app/main", "-c"]