# セキュリティ設定ガイド

このドキュメントは、cinetagフロントエンドのセキュリティ設定について説明します。

## 実装済みセキュリティヘッダー

### Content Security Policy (CSP)

`next.config.ts` で以下のCSPヘッダーを設定しています。

#### 主要なディレクティブ

| ディレクティブ | 設定値 | 目的 |
|--------------|--------|------|
| `default-src` | `'self'` | デフォルトで同一オリジンのみ許可 |
| `script-src` | `'self' 'unsafe-inline' 'unsafe-eval'` | Next.js/React の動作に必要な設定 |
| `style-src` | `'self' 'unsafe-inline' https://fonts.googleapis.com` | インラインスタイルとGoogle Fonts |
| `img-src` | `'self' data: https://placehold.co https://image.tmdb.org https://img.clerk.com https://images.clerk.dev` | 画像ソースのホワイトリスト |
| `connect-src` | 環境依存 | APIエンドポイントのホワイトリスト |
| `frame-src` | `'self' https://clerk.com https://*.clerk.accounts.dev` | Clerk認証用iframe |
| `object-src` | `'none'` | Flash等のプラグイン禁止 |
| `base-uri` | `'self'` | 相対URLハイジャック対策 |
| `form-action` | `'self'` | フォーム送信先を自サイトのみに制限 |
| `frame-ancestors` | `'none'` | クリックジャッキング対策 |

#### 環境別設定

- **開発環境**: `connect-src` に `http://localhost:8080` を含む
- **本番環境**: `upgrade-insecure-requests` を有効化してHTTPをHTTPSに自動変換

### その他のセキュリティヘッダー

| ヘッダー | 設定値 | 目的 |
|---------|--------|------|
| `X-Content-Type-Options` | `nosniff` | MIMEタイプスニッフィング防止 |
| `X-Frame-Options` | `DENY` | クリックジャッキング対策 |
| `X-XSS-Protection` | `1; mode=block` | 旧ブラウザ向けXSS対策 |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | リファラー情報の制御 |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=(), interest-cohort=()` | 不要な機能の無効化 |

## CSP違反のモニタリング

### レポートエンドポイント

CSP違反は `/api/csp-report` エンドポイントで受信されます。

- **開発環境**: コンソールに警告を出力
- **本番環境**: ログ収集サービス（Sentry等）への送信を推奨

### 違反レポートの確認方法

#### ブラウザのDevToolsで確認

```javascript
// ブラウザのコンソールで実行
fetch(window.location.href)
  .then(res => {
    console.log('CSP Header:', res.headers.get('content-security-policy'));
  });
```

CSP違反が発生すると、コンソールに以下のようなエラーが表示されます:

```
Refused to load the script 'https://example.com/malicious.js'
because it violates the following Content Security Policy directive:
"script-src 'self' 'unsafe-inline' 'unsafe-eval'".
```

#### サーバーログで確認

開発サーバーのログに `🚨 CSP Violation Report:` という形式で出力されます。

## テスト手順

### 1. CSPヘッダーが正しく設定されているか確認

```bash
# 開発サーバーを起動
cd apps/frontend
npm run dev

# 別のターミナルで
curl -I http://localhost:3000 | grep -i "content-security-policy"
```

期待される出力:
```
content-security-policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...
```

### 2. 主要機能のテスト

以下の機能が正常に動作することを確認してください:

- ✅ ページの読み込み
- ✅ Clerk認証（サインイン/サインアウト）
- ✅ 画像の表示（TMDB、Clerk、placehold.co）
- ✅ API通信（タグ一覧、詳細取得など）
- ✅ モーダル表示
- ✅ Google Fonts の読み込み

### 3. CSP違反のシミュレーション

意図的にCSP違反を発生させてレポートが正しく送信されるかテスト:

```javascript
// ブラウザのコンソールで実行
const script = document.createElement('script');
script.src = 'https://evil.example.com/malicious.js';
document.body.appendChild(script);
```

期待される動作:
- コンソールにCSP違反のエラーが表示される
- サーバーログに違反レポートが記録される（開発環境）

## 本番環境への展開

### 1. 環境変数の設定

本番環境のAPIエンドポイントを確認し、`next.config.ts` の `connectSrc` を更新してください。

```typescript
const connectSrc = isDev
  ? "'self' https://clerk.com https://*.clerk.accounts.dev http://localhost:8080"
  : "'self' https://clerk.com https://*.clerk.accounts.dev https://api.cinetag.com"; // 本番APIのURL
```

### 2. ログ収集サービスの設定

`apps/frontend/src/app/api/csp-report/route.ts` にログ収集サービスへの送信処理を追加してください。

例: Sentryの場合

```typescript
import * as Sentry from '@sentry/nextjs';

// POST関数内
if (process.env.NODE_ENV === 'production') {
  Sentry.captureMessage('CSP Violation', {
    level: 'warning',
    extra: report,
    tags: {
      type: 'csp_violation',
    },
  });
}
```

### 3. デプロイ前のチェックリスト

- [ ] `next.config.ts` の `connectSrc` に本番APIのURLを追加
- [ ] ステージング環境でCSPをテスト
- [ ] 全ての主要機能が動作することを確認
- [ ] CSP違反レポートの収集先を設定
- [ ] HTTPS証明書が有効であることを確認

## 将来の改善予定

### Nonceベースの厳格なCSP

現在は `'unsafe-inline'` と `'unsafe-eval'` を許可していますが、将来的にはNonceベースのCSPに移行することを推奨します。

メリット:
- インラインスクリプト攻撃の完全防止
- より厳格なセキュリティ

実装方法は `apps/frontend/docs/csp-nonce-implementation.md` を参照してください。

## トラブルシューティング

### 問題: Clerkの認証モーダルが表示されない

**原因**: `frame-src` にClerkのドメインが含まれていない

**解決策**: `next.config.ts` で以下を確認

```typescript
"frame-src 'self' https://clerk.com https://*.clerk.accounts.dev"
```

### 問題: Google Fontsが読み込まれない

**原因**: `style-src` または `font-src` が不足

**解決策**: 以下のディレクティブを確認

```typescript
"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com"
"font-src 'self' https://fonts.gstatic.com data:"
```

### 問題: TMDB画像が表示されない

**原因**: `img-src` にTMDBのドメインが含まれていない

**解決策**: 以下を確認

```typescript
"img-src 'self' data: https://placehold.co https://image.tmdb.org https://img.clerk.com https://images.clerk.dev"
```

## 参考資料

- [MDN - Content Security Policy](https://developer.mozilla.org/ja/docs/Web/HTTP/CSP)
- [Next.js - Content Security Policy](https://nextjs.org/docs/app/building-your-application/configuring/content-security-policy)
- [OWASP - CSP Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
- [CSP Evaluator](https://csp-evaluator.withgoogle.com/)

## お問い合わせ

CSPに関する問題や質問がある場合は、開発チームまでご連絡ください。
