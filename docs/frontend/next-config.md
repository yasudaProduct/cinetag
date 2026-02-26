# next.config.ts 設定ガイド

## 概要

`apps/frontend/next.config.ts` は Next.js の設定ファイルで、以下の3つの責務を担う。

1. **React Compiler の有効化**
2. **外部画像ホストの許可リスト**（`next/image` 用）
3. **セキュリティレスポンスヘッダーの定義**（CSP 含む）

---

## 1. React Compiler

```ts
reactCompiler: true
```

React 19 の React Compiler を有効化。`useMemo` / `useCallback` による手動メモ化をコンパイラが自動で行う。

---

## 2. 外部画像ホスト許可（`images.remotePatterns`）

`next/image` コンポーネントで表示を許可する外部ホストの定義。

| ホスト | 用途 |
|---|---|
| `placehold.co` | プレースホルダー画像 |
| `image.tmdb.org` | TMDB の映画ポスター・バックドロップ |
| `img.clerk.com` | Clerk ユーザーアバター（本番） |
| `images.clerk.dev` | Clerk ユーザーアバター（開発） |

### 変更が必要になるケース

- 新しい外部画像サービスを利用する場合 → ホストを追加
- Clerk がドメインを変更した場合 → ホストを更新

---

## 3. セキュリティレスポンスヘッダー

`headers()` 関数で全ルート（`/:path*`）に付与。開発/本番で値が動的に切り替わる。

### 3.1 環境依存の動的値

設定に使用される環境変数とその用途:

| 環境変数 | 用途 |
|---|---|
| `NODE_ENV` | 開発/本番の判定（`isDev`） |
| `NEXT_PUBLIC_BACKEND_API_BASE` | `connect-src` にバックエンドオリジンを追加 |
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | キーのプレフィックスで Clerk ホスト（`.com` / `.dev`）を判定 |

**Clerk ホストの判定ロジック:**

| Publishable Key のプレフィックス | 許可するホスト |
|---|---|
| `pk_live_` | `*.clerk.accounts.com` |
| `pk_test_` | `*.clerk.accounts.dev` |
| 上記以外（開発） | `*.clerk.accounts.dev` |
| 上記以外（本番） | 両方 |

### 3.2 Content Security Policy（CSP）

各ディレクティブの許可対象と、その理由:

| ディレクティブ | 許可対象 | 理由 |
|---|---|---|
| `default-src` | `'self'` | 基本ポリシー。明示的に許可していないリソースは同一オリジンのみ |
| `script-src` | `'self'`, `'unsafe-inline'`, `'unsafe-eval'`, Clerk ホスト, `clerk.cine-tag.com`, Cloudflare Insights※ | Next.js / React がインラインスクリプトと eval を使用。Clerk SDK のスクリプト読み込み |
| `style-src` | `'self'`, `'unsafe-inline'`, `fonts.googleapis.com` | インラインスタイル（Tailwind CSS / shadcn/ui）と Google Fonts |
| `font-src` | `'self'`, `fonts.gstatic.com`, `data:` | Google Fonts のフォントファイル |
| `img-src` | `'self'`, `data:`, `blob:`, `placehold.co`, `image.tmdb.org`, `img.clerk.com`, `images.clerk.dev` | 外部画像サービス。`data:` / `blob:` は Next.js の画像最適化で使用 |
| `connect-src` | `'self'`, `clerk.com`, Clerk ホスト, バックエンドオリジン, `localhost:8080`※, Cloudflare Insights※ | fetch / XHR の接続先 |
| `worker-src` | `'self'`, `blob:` | Web Worker 生成（blob URL 経由） |
| `frame-src` | `'self'`, `clerk.com`, Clerk ホスト | Clerk の認証モーダル（iframe） |
| `object-src` | `'none'` | Flash 等のプラグインを完全禁止 |
| `base-uri` | `'self'` | `<base>` タグによる相対 URL ハイジャック防止 |
| `form-action` | `'self'` | フォーム送信先を自サイトに制限 |
| `frame-ancestors` | `'none'` | 他サイトでの iframe 埋め込み禁止（クリックジャッキング対策） |
| `upgrade-insecure-requests` | （本番のみ） | HTTP を HTTPS に自動昇格 |

※ 開発/本番で異なる:
- `localhost:8080` → 開発のみ
- `static.cloudflareinsights.com`（script-src）/ `cloudflareinsights.com`（connect-src） → 本番のみ

### 3.3 その他のセキュリティヘッダー

| ヘッダー | 値 | 目的 |
|---|---|---|
| `X-Content-Type-Options` | `nosniff` | MIME タイプスニッフィング防止 |
| `X-Frame-Options` | `DENY` | クリックジャッキング対策（CSP `frame-ancestors` との二重防御） |
| `X-XSS-Protection` | `1; mode=block` | 旧ブラウザ向け XSS フィルター |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | 同一オリジンではフル URL、クロスオリジンではオリジンのみ送信 |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | 不要なブラウザ API を無効化 |

---

## 変更が必要になるケース一覧

| シナリオ | 変更箇所 |
|---|---|
| 新しい外部画像サービスを使う | `images.remotePatterns` にホストを追加 |
| 新しい外部スクリプト（Analytics等）を導入する | `script-src` にスクリプト配信元のドメインを追加 |
| 新しい外部 API に fetch する | `connect-src` に API のオリジンを追加 |
| 外部フォントサービスを追加する | `style-src`（CSS）と `font-src`（フォントファイル）にドメインを追加 |
| iframe で外部サービスを埋め込む | `frame-src` にドメインを追加 |
| Clerk のカスタムドメインを変更する | `script-src` の `clerk.cine-tag.com` を更新 |
| バックエンド API のオリジンが変わる | 環境変数 `NEXT_PUBLIC_BACKEND_API_BASE` を更新（自動反映） |
| Clerk の Publishable Key が変わる | 環境変数 `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` を更新（自動反映） |
