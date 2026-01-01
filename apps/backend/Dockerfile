FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./src/cmd

# 実行用
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# ポート8080を公開（Cloudflare Containersのデフォルト）
EXPOSE 8080

# 環境変数PORTを設定可能にする
ENV PORT=8080

# アプリケーションを起動
CMD ["./main"]