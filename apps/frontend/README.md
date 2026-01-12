# cinetag frontend（Next.js / React）

`cinetag` のフロントエンド実装です。Next.js 16（App Router）と React 19 を使用したモダンな Web アプリケーションです。

---

## 技術スタック

- **フレームワーク**: Next.js 16.0.10 with App Router
- **React**: 19.2.0（React Compiler有効）
- **TypeScript**: 5.x
- **状態管理**: TanStack React Query 5.x
- **認証**: Clerk (@clerk/nextjs)
- **スタイリング**: Tailwind CSS v4 + PostCSS
- **UIコンポーネント**: shadcn/ui + Radix UI
- **バリデーション**: Zod
- **アイコン**: Lucide React
- **デプロイ**: Cloudflare Pages（@opennextjs/cloudflare）

詳細なアーキテクチャやAPIレイヤーの設計は `docs/frontend/frontend-api-layer.md` を参照してください。

---

## ディレクトリ構成

```text
apps/frontend/src/
├── app/                          # Next.js App Router
│   ├── (auth)/               # 認証ルート（ルートグループ）
│   │   ├── sign-in/          # サインイン
│   │   └── sign-up/          # サインアップ
│   ├── [username]/           # ユーザーページ
│   ├── tags/                 # タグ関連
│   │   └── [tagId]/          # タグ詳細
│   ├── mypage/               # マイページ
│   └── layout.tsx             # ルートレイアウト（プロバイダー含む）
├── components/
│   ├── providers/            # React Query プロバイダーなど
│   └── ui/                   # shadcn/ui コンポーネント
└── lib/
    ├── api/                  # API レイヤー（リソース別）
    │   ├── _shared/          # 共通ユーティリティ（http.ts, auth.ts）
    │   ├── tags/             # タグAPI関数
    │   ├── movies/            # 映画API関数
    │   └── users/             # ユーザーAPI関数
    ├── validation/            # Zod スキーマ
    └── mock/                 # 開発用モックデータ
```

より詳細な責務分担についても `docs/frontend/frontend-api-layer.md` を参照してください。

---

## 開発環境の準備

### 1. 依存パッケージのインストール

```bash
cd apps/frontend
npm install
```

### 2. 環境変数の設定

`.env.example` ファイルをコピーして `.env.local` ファイルを作成し、必要な環境変数を設定します:

```bash
cd apps/frontend
cp .env.example .env.local
```

`.env.local` ファイルを編集して、以下の環境変数を設定してください:

- `NEXT_PUBLIC_BACKEND_API_BASE` - バックエンドAPIのベースURL（例: `http://localhost:8080`）
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` - Clerk公開キー
- `CLERK_SECRET_KEY` - Clerkシークレットキー

### 3. 開発サーバーの起動

```bash
cd apps/frontend
npm run dev
```

フロントエンドは `http://localhost:3000` で起動します。

---

## よく使うコマンド

```bash
# 開発サーバーの起動
npm run dev

# プロダクションビルド
npm run build

# プロダクションサーバーの起動
npm start

# リンターの実行
npm run lint

# Cloudflare Pages 用のビルドとプレビュー
npm run preview

# Cloudflare Pages へのデプロイ
npm run deploy
```

---

## APIレイヤーのパターン

すべてのAPI呼び出しは以下のパターンに従います:

1. `lib/api/_shared/http.ts` の集約されたfetchユーティリティを使用
2. `lib/validation/` のZodスキーマでレスポンスを検証
3. React Query（`useQuery`, `useMutation`）でサーバー状態を管理
4. `lib/api/_shared/auth.ts` の `getBackendTokenOrThrow()` で認証トークンを処理

例:

```typescript
// コンポーネント内
const { data } = useQuery({
  queryKey: ["tags"],
  queryFn: listTags
});

// lib/api/tags/list.ts 内
export async function listTags(): Promise<TagsList> {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags`);
  const body = await safeJson(res);
  if (!res.ok) throw new Error(toApiErrorMessage({...}));
  return TagsListResponseSchema.safeParse(body).data.items;
}
```

詳細は `docs/frontend/frontend-api-layer.md` を参照してください。

---

## 認証

- `middleware.ts` の `clerkMiddleware()` によるルート保護
- 公開ルート: `/`, `/sign-in`, `/sign-up`, `/auth/test-signin`
- トークン注入: 認証が必要なAPI呼び出しには `getBackendTokenOrThrow()` を使用
- Clerkテンプレート名: "cinetag-backend"

詳細は `docs/architecture/auth-architecture.md` を参照してください。

---

## スタイリング方針

- **ユーティリティファースト**: インラインTailwindクラス
- **CSS変数**: `:root` にoklch色空間を使用したテーマカラー
- **コンポーネントバリアント**: class-variance-authority (CVA) を使用
- **ダークモード**: CSS変数でサポート

---

## 環境変数

**フロントエンド** (`apps/frontend/.env.local`):

- `NEXT_PUBLIC_BACKEND_API_BASE` - バックエンドAPIのURL（必須）
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` - Clerk公開キー（必須）
- `CLERK_SECRET_KEY` - Clerkシークレットキー（必須）

---

## デプロイ

### Cloudflare Pages

このプロジェクトは Cloudflare Pages へのデプロイを想定しています。

#### ローカルからのデプロイ

```bash
# ビルドとプレビュー
npm run preview

# デプロイ
npm run deploy
```

#### GitHub Actions からの自動デプロイ

GitHub Actions を導入する場合、`develop` ブランチへの push をトリガーにして Cloudflare（Pages/Workers）へ自動デプロイできます。

##### 設定手順

1. **Cloudflare API Token の作成**

   - Cloudflare ダッシュボード → My Profile → API Tokens
   - 「Create Token」をクリック
   - 「Edit Cloudflare Workers」テンプレートを使用、またはカスタムトークンを作成
   - 必要な権限:
     - Account: Cloudflare Workers:Edit
     - Zone: Zone Settings:Read, Zone:Read
   - トークンをコピー（一度しか表示されません）

2. **GitHub Secrets の設定**

   GitHub リポジトリの Settings → Secrets and variables → Actions で、以下のシークレットを追加:

   - `CLOUDFLARE_API_TOKEN`: Cloudflare API Token（必須）
   - `NEXT_PUBLIC_BACKEND_API_BASE`: バックエンドAPIのURL（例: `https://api.example.com`）
   - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`: Clerk公開キー
   - `CLERK_SECRET_KEY`: Clerkシークレットキー

3. **ワークフローの動作**

   ワークフロー（例: `.github/workflows/ci-develop.yml`）の `frontend-deploy` ジョブで以下を実行:
   - 依存パッケージのインストール
   - `npm run deploy` によるビルドとデプロイ
   - 環境変数をビルド時に注入

4. **実行タイミング**

   - `develop` ブランチへの push 時に自動実行
   - 他のCIジョブ（テスト、マイグレーションなど）と並列実行

詳細は `docs/architecture/infrastructure-configuration.md` を参照してください。

> 補足: 既存のワークフローは `/.github/workflows/ci-develop.yml` を参照してください。CI/CD全体の方針は `docs/operations/cicd.md` にまとめています。

---

## API 仕様

バックエンドAPIの詳細は `docs/api/api-spec.md` を参照してください。

---

## 関連ドキュメント

- バックエンドの詳細: [apps/backend/README.md](../backend/README.md)
- API仕様: [docs/api/api-spec.md](../../docs/api/api-spec.md)
- フロントエンドAPIレイヤー設計: [docs/frontend/frontend-api-layer.md](../../docs/frontend/frontend-api-layer.md)
- 認証アーキテクチャ: [docs/architecture/auth-architecture.md](../../docs/architecture/auth-architecture.md)
- バリデーション: [docs/frontend/frontend-validation.md](../../docs/frontend/frontend-validation.md)
