# フロントエンド API レイヤー設計（`apps/frontend/src/lib/api`）

## 目的

`apps/frontend/src/lib/api` は、フロントエンドからバックエンド API を呼び出すための **APIアクセス層（低レベルHTTPラッパ）** を提供する。

- UI（コンポーネント）から **URL・HTTPメソッド・ヘッダ・body・レスポンス検証・エラー整形** を隔離する
- バックエンド仕様の揺れや想定外レスポンスで UI が壊れるのを防ぐ（Runtime Validation）
- 認証が必要なAPI呼び出しの「トークンの受け渡し」を標準化する

関連ドキュメント：

- `docs/api/api-spec.md`（バックエンドAPI仕様）
- `docs/architecture/auth-architecture.md`（Clerk認証設計）
- `docs/frontend/frontend-validation.md`（zodによる入力/レスポンス検証ルール）

---

## `lib/api` 配下の責務

### やること（責務）

- **エンドポイント単位の関数提供**
  - 例：`listTags()` / `createTag()` / `getTag()` / `listTagMovies()` など
- **HTTP詳細のカプセル化**
  - URL組み立て、`fetch` オプション、ヘッダ、body、`encodeURIComponent` 等
- **レスポンスの runtime validation（zod `safeParse`）**
  - `fetch().json()` は `unknown` として扱い、Schemaで検証する（`docs/frontend/frontend-validation.md`）
- **APIエラーの解釈・例外化**
  - 例：`{"error":"..."}` 形式ならその文言を優先し、なければ `...（status）` の一般メッセージ
- **フロント向けの正規化（必要な範囲）**
  - バックエンドの差異を吸収し、UIが扱いやすい形に整える（ただし“UI都合の派手な整形”は避ける）

### やらないこと（禁止/非責務）

- **React/React Query に依存するコード**
  - `useQuery` / `useMutation` を `lib/api` に置かない（HookはUI層または専用hooks層へ）
- **コンポーネント/表示ロジック**
  - UI文言、state管理、表示条件、DOM操作などは置かない
- **フォーム入力バリデーション（送信前）**
  - 入力検証は `lib/validation/*` とUI側（フォーム）で行う

---

## ファイル分割ルール

### 基本方針

- ルール：**リソース名 = 複数形ディレクトリ**
  - `tags/`, `users/`, `movies/` …
- `index.ts` は公開API（exportする関数）だけをまとめる（barrel）

### 推奨ディレクトリ例

```text
apps/frontend/src/lib/api/
  tags/
    index.ts
    list.ts          # GET    /api/v1/tags
    create.ts        # POST   /api/v1/tags
    detail.ts        # GET    /api/v1/tags/:tagId
    movies.ts        # GET    /api/v1/tags/:tagId/movies
  users/
    index.ts
    me.ts            # GET    /api/v1/me 等（設計次第で me/ に分離してもOK）
```

### 命名ルール（推奨）

- できるだけ **HTTPやUIに引きずられない**動詞 + 名詞に寄せる
  - `listTags`, `createTag`, `getTag`, `listTagMovies`
- 同じリソースでは命名を揃える（`listX / getX / createX / updateX / deleteX`）

---

## 実行環境（Client/Server）と分割方針

Next.js（App Router）では、**呼び出し元が Client Component か Server Component か**で、同じ関数でも実行環境が変わる。

- Client Component（先頭に `"use client"`）から呼ばれる → **ブラウザ実行**
- Server Component / Route Handler / Server Action から呼ばれる → **サーバー実行**

この差を安全に扱うため、`lib/api` は原則「両対応」に寄せ、サーバー専用が必要になったら明示的に隔離する。

---

## `lib/api/**` は原則 Client/Server 両対応 (isomorphic)

### 目的

`lib/api/**` を「どこから import しても動く」状態にすることで、UI（クライアント）でもServer Componentでも同じAPI関数を使えるようにする。

### 守ること（仕様）

- **依存は `fetch` と共通ユーティリティ（zod等）に限定**する
- **秘密情報（server-only env）に依存しない**
  - 参照する env は原則 `NEXT_PUBLIC_*` のみ
- **認証が必要な場合、トークンは引数で受け取る**
  - 例：`createTag({ token, input })`
  - トークン取得（Clerk）は呼び出し側で行う（Client: `useAuth().getToken()` など）
- **レスポンス検証は zod `safeParse` を標準**
  - `fetch().json()` → `unknown` → schemaで検証 → 失敗時は `console.warn` + 例外
  - 詳細は `docs/frontend/frontend-validation.md` に従う
- **APIエラーは共通形式を優先して解釈**
  - `{"error":"..."}` が取れるならその文言を使う（取れないなら status を含む一般メッセージ）

### やってはいけない（NG）

- `@clerk/nextjs/server` や `next/headers` 等の **server-only API** を使う
- `window` / `document` 等の **ブラウザ専用API** を使う
- `fs` 等の **Node専用API** を使う
- React Hook（`useQuery`, `useMutation`）を置く

---

## サーバー専用が必要になったら `lib/api/server/**` に隔離

### 使う場面

以下の要件が出たら、isomorphic ではなく server-only に寄せる：

- バックエンド直叩きを避け、サーバー経由で呼びたい（CORS/秘匿/統制）
- Cookie/セッションなど **サーバーでしか扱えない情報** を使う
- `NEXT_PUBLIC` に出せない **秘密env** が必要
- Clerkの server SDK でトークン取得・検証をしたい

### 仕様（守ること）

- 置き場所：`apps/frontend/src/lib/api/server/*`
- **server-only を宣言**して、誤ってクライアントから import できないようにする
  - ファイル先頭に `import "server-only";` を置く（推奨）
- Route Handler / Server Action / Server Component からのみ呼ぶ
- `@clerk/nextjs/server` や `next/headers` を使ってよい（必要な範囲で）

---

## 実装上の共通ルール（推奨）

### 1) 共通ユーティリティの集約

`safeJson` / baseURL取得 / エラー整形 / fetchオプションの共通化は、重複と揺れを減らすために推奨。

例：

- `lib/api/_shared/http.ts`（isomorphic）
  - `safeJson`
  - `getBaseOrThrow`（NEXT_PUBLIC 前提）
  - `parseApiError`（`ApiErrorSchema`）
  - `parseApiError`（`ApiErrorSchema`）

※ `_shared` のような「リソース非依存の共通部」は、`lib/api` 直下にまとめる。

### 2) 型/スキーマの置き場所

- Schema：`apps/frontend/src/lib/validation/*` に集約（`docs/frontend/frontend-validation.md`）
- API関数の入出力型：スキーマから `z.infer` で導出するか、最小限の型のみ `lib/api` 内に置く

---

## 移行メモ（現状の `tag.ts` / `tags.ts` から）

- `apps/frontend/src/lib/api/tags.ts`（一覧/作成）と `apps/frontend/src/lib/api/tag.ts`（詳細/配下）を、
  `apps/frontend/src/lib/api/tags/*` に統合する（案1）
- 既存の「モックフォールバック」等の扱いは、最終方針を決めてから統一する
  - 例：base未設定時は常にthrowに統一する / 開発環境のみモック許容に統一する、など


