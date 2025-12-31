# マルチステージビルド: ビルドステージ
FROM golang:1.23-alpine AS builder

# 必要なビルドツールをインストール
RUN apk add --no-cache git

WORKDIR /app

# 依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# 実行ステージ
FROM alpine:latest

# 実行に必要なパッケージをインストール
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server .

# ポート8080を公開
EXPOSE 8080

# サーバーを起動
CMD ["./server"]
