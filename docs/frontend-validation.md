# フロントエンド バリデーション仕様（zod）

## 目的

- **入力バリデーション（UI/UX）**: ユーザーが送信前に問題へ気づけるようにする
- **APIレスポンスの実行時検証（Runtime Validation）**: バックエンドの変更や想定外データで UI が壊れるのを防ぐ

> 注意: **最終的な正はバックエンド**。フロントのバリデーションは「UX向上」と「壊れにくさ（耐障害性）」が目的。

## 採用ライブラリ

- **zod**: `apps/frontend/package.json` の `dependencies` に含む

## 実装方針（ルール）

### 1) フォーム送信前は必ず `safeParse`

- `parse()` は例外を投げるため、UIコードでは **`safeParse()` を標準**にする
- エラー時は zod のエラーメッセージ（日本語）を UI に表示する

### 2) APIレスポンスは必ず `safeParse`

- `fetch().json()` の結果は `unknown` として扱い、zod で検証する
- 検証NGの場合:
  - 画面をクラッシュさせない（例: `setTags([])` などの安全なフォールバック）
  - `console.warn` に **検証エラーとボディ**を出して調査可能にする

### 3) スキーマは「共通モジュール」に集約

- 置き場所: `apps/frontend/src/lib/validation/`
- 原則: **コンポーネントにスキーマを散らさない**

## タグ関連スキーマ（v1）

### ファイル

- `apps/frontend/src/lib/validation/tag.ts`

### タグ作成（入力）

- 対象UI: `TagCreateModal`
- ルール（バックエンド仕様と合わせる）:
  - `title`: **1〜100文字**
  - `description`: **最大500文字**（未入力は省略）

### タグ一覧取得（レスポンス）

- 対象API: `GET /api/v1/tags`
- ドキュメント上の基本形: `{ items: Tag[] }`
- 実装差分に備え、フロントでは **配列のみのレスポンス**も許容し `{ items }` へ正規化する

### タグ作成（レスポンス）

- 対象API: `POST /api/v1/tags`
- 成功レスポンスを検証し、想定外の形なら **UI では「形式が不正」扱い**にする

## エラー表示仕様（暫定）

- 入力エラー: zod の message をそのまま出す（日本語メッセージをスキーマに持たせる）
- API エラー:
  - `{"error": "..."}` 形式ならその `error` を表示
  - それ以外は `...（status）` の一般メッセージ

## 今後の拡張方針

- API スキーマ（OpenAPI/Swagger）が整備されてきたら:
  - **OpenAPI → TypeScript 型生成**（例: `openapi-typescript`）
  - さらに必要なら **型 + zod の整合**（生成 or 手書き）を検討
- フォームが増えたら:
  - React Hook Form + zod resolver の導入でフォーム実装を共通化する


