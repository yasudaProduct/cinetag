## cinetag backend（Go / Gin）

`cinetag` のバックエンド API 実装です。フロントエンド（Next.js, `apps/frontend`）からのリクエストを受け、映画・タグなどのドメインを扱う REST API を提供します。

---

## 技術スタック

- **言語**: Go (Go 1.25 系)
- **Web フレームワーク**: Gin (`github.com/gin-gonic/gin`)
- **ORM**: GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- **DB**: PostgreSQL（開発環境では Docker を想定）

詳細なアーキテクチャやレイヤー構成は `docs/backend-architecture.md` を参照してください。

---

## ディレクトリ構成（抜粋）

```text
apps/backend/
├── src/
│   ├── cmd/
│   │   ├── main.go              # API サーバーのエントリーポイント
│   │   └── migrate/
│   │       └── main.go          # DB マイグレーション用コマンド
│   ├── internal/
│   │   ├── handler/             # HTTP ハンドラー
│   │   ├── service/             # ビジネスロジック
│   │   ├── model/               # ドメインモデル
│   │   └── middleware/          # カスタムミドルウェア
│   └── router/
│       └── router.go            # ルーティング定義
├── go.mod
└── go.sum
```

より詳細な責務分担についても `docs/backend-architecture.md` を参照してください。

---

## 開発環境の準備

### 1. 依存パッケージのインストール

プロジェクトルートで以下を実行します（`apps/backend` 配下でも可）:

```bash
cd apps/backend
go mod tidy
```

### 2. データベース（PostgreSQL）の起動

プロジェクトルートにある `compose.yml` を利用して PostgreSQL を起動します:

```bash
cd /Users/yuta/Develop/cinetag
docker compose up -d postgres
```

> ポート `5432` で `postgres/postgres` ユーザー・パスワードの DB が立ち上がります。

---

## サーバーの起動

`apps/backend` ディレクトリで以下を実行します:

```bash
cd apps/backend
go run ./src/cmd
```

デフォルトではポート `8080` で起動する想定です（実際のポートや環境変数の仕様は `cmd/main.go` を参照してください）。

---

## 認証（Clerk JWT 検証）に必要な環境変数

- **`CLERK_JWKS_URL`（必須）**: Clerk の JWKS エンドポイント（例: `https://<your-domain>/.well-known/jwks.json`）
- **`CLERK_ISSUER`（任意）**: 期待する `iss`（設定時のみ検証します）
- **`CLERK_AUDIENCE`（任意）**: 期待する `aud`（設定時のみ検証します）

---

## ngrok を使った公開 URL の作成（例: 外部 Webhook 連携）

Clerk などの外部サービスからローカル開発環境に Webhook を受けたい場合は、`ngrok` を使ってローカルの 8080 ポートをインターネットに公開します。

1. **ngrok で 8080 を公開**

   別ターミナルで以下を実行します:

   ```bash
   ngrok http 8080
   ```

2. **発行された HTTPS URL を外部サービスに設定**

   - ngrok のコンソールに表示される `https://xxxxx.ngrok.io` のような URL をコピーし、
   - Clerk やその他 Webhook 送信元の「エンドポイント URL」として設定します  
     （例: `https://xxxxx.ngrok.io/api/webhooks/clerk` など）

ngrok 経由の URL はセッションごとに変わるため、**ngrok を再起動した場合は Webhook 設定側の URL も更新**してください。

---

## API 仕様

提供しているエンドポイントやレスポンス形式などの詳細は、`docs/api-spec.md` を参照してください。

---

## 今後の拡張メモ

- `internal/config` を追加し、環境変数・設定値の読み込みを集約する
- ログ出力の統一（`log` / `slog` など）とログフォーマットの整理
- 認証・認可（Clerk など）まわりのハンドラー・ミドルウェアの充実


