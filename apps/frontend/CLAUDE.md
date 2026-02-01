# CLAUDE.md - Frontend

Next.js + Reactフロントエンドのガイダンスです。

## 技術スタック

- **フレームワーク**: Next.js 16.0.5 with App Router
- **React**: 19.2.0（React Compiler有効）
- **TypeScript**: 5.x
- **状態管理**: TanStack React Query 5.x
- **認証**: Clerk (@clerk/nextjs)
- **スタイリング**: Tailwind CSS v4 + PostCSS
- **UIコンポーネント**: shadcn/ui + Radix UI
- **バリデーション**: Zod
- **アイコン**: Lucide React

## ディレクトリ構成

```
src/
├── app/                    # Next.js App Router
│   ├── (auth)/            # 認証ルート（ルートグループ）
│   ├── tags/[tagId]/      # 動的ルート
│   └── layout.tsx         # プロバイダー含むルートレイアウト
├── components/
│   ├── providers/         # React Queryなど
│   └── ui/               # shadcn/uiコンポーネント
└── lib/
    ├── api/              # リソース別に整理されたAPIレイヤー
    │   ├── _shared/      # http.ts（fetchユーティリティ）、auth.ts（トークン）
    │   ├── tags/         # タグAPI関数
    │   └── movies/       # 映画API関数
    ├── validation/       # Zodスキーマ
    └── mock/            # 開発用モックデータ
```

## よく使うコマンド

```bash
# 依存パッケージのインストール
npm install

# 開発サーバーの起動
npm run dev

# プロダクションビルド
npm run build

# プロダクションサーバーの起動
npm start

# リンターの実行
npm run lint
```

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

## 認証

- `middleware.ts` の `clerkMiddleware()` によるルート保護
- 公開ルート: `/`, `/sign-in`, `/sign-up`
- トークン注入: 認証が必要なAPI呼び出しには `getBackendTokenOrThrow()` を使用
- Clerkテンプレート名: "cinetag-backend"

## スタイリング方針

- **ユーティリティファースト**: インラインTailwindクラス
- **CSS変数**: `:root` にoklch色空間を使用したテーマカラー
- **コンポーネントバリアント**: class-variance-authority (CVA) を使用
- **ダークモード**: CSS変数でサポート

## 環境変数

`.env.local` ファイルに設定:
- `NEXT_PUBLIC_BACKEND_API_BASE` - バックエンドAPI URL
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` - Clerk公開キー
- `CLERK_SECRET_KEY` - Clerkシークレットキー

## テスト戦略

- React Query DevToolsでの手動テスト
- リンティング: `npm run lint`

## フロントエンドAPI統合の追加

1. `lib/validation/` でZodスキーマを定義
2. `lib/api/{resource}/` でAPI関数を作成
3. コンポーネントで適切なクエリキーを使用してReact Queryを利用
4. UIでローディング/エラー状態を処理
