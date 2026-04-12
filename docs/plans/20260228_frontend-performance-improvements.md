# フロントエンド パフォーマンス改善

**作成日**: 2026-02-28
**ブランチ**: `feature/frontend-performance-improvements`
**対象イシュー**: #45, #46, #47, #48, #49
**PR**: #50
**ステータス**: 完了

---

## 1. 背景と目的

Cloudflare Workers（Freeプラン）上で動作するNext.jsフロントエンドのパフォーマンスを改善する。
主に以下の観点から最適化を行った:

- **ISRキャッシュの永続化**: R2バケットによるインクリメンタルキャッシュ
- **不要なランタイムCPU消費の回避**: Cloudflare Freeプランの10ms CPU制限を考慮
- **初期バンドルサイズの削減**: 遅延ロードによるモーダルコンポーネントの分離
- **動的ページのレスポンス改善**: ISR導入によるキャッシュ活用

---

## 2. 各イシューの対応内容

### 2.1 #45: ISRキャッシュのR2永続化

**課題**: ISRで生成されたキャッシュがWorkerの再デプロイ時に消失する

**対応ファイル**:

| ファイル | 変更内容 |
|---------|---------|
| `apps/frontend/wrangler.jsonc` | R2バケット設定のコメントアウトを解除 |
| `apps/frontend/open-next.config.ts` | `r2IncrementalCache` をインクリメンタルキャッシュとして設定 |

**wrangler.jsonc の変更**:
```jsonc
"r2_buckets": [
  {
    "binding": "NEXT_INC_CACHE_R2_BUCKET",
    "bucket_name": "cache"
  }
],
```

**open-next.config.ts の変更**:
```typescript
import { defineCloudflareConfig } from "@opennextjs/cloudflare/config";
import r2IncrementalCache from "@opennextjs/cloudflare/overrides/incremental-cache/r2-incremental-cache";

export default defineCloudflareConfig({
  incrementalCache: r2IncrementalCache,
});
```

**デプロイ時の前提条件**:
- 事前に `wrangler r2 bucket create cache` でR2バケットの作成が必要

---

### 2.2 #46: ホームページのデータ取得方針

**課題**: ホームページのタグ一覧データ取得方式の最適化

**検討の経緯**:

1. イシューではクライアントサイドフェッチ（React Query + スケルトンUI）への変更を提案
2. しかし元々の実装はSSGで、ビルド時にデータを取得してランタイムではデータ取得を行わない設計
3. Cloudflare Freeプランの10ms CPU制限を考慮し、各方式を比較検討

**方式比較**:

| 方式 | 初回表示速度 | ランタイムCPU | データ鮮度 |
|------|------------|-------------|----------|
| 純粋SSG（`revalidate`なし） | 最速（静的HTMLを返すのみ） | ゼロ | ビルド時のまま |
| ISR（`revalidate`あり） | 速い（キャッシュから返す） | 再生成時にReactレンダリング消費 | 定期更新 |
| クライアントサイドフェッチ | 遅い（追加APIラウンドトリップ） | ゼロ | リアルタイム |

**最終方針**: 純粋SSG（`revalidate`を設定しない）

**理由**:
- ホームページのタグ一覧はリアルタイム性を必要としない
- 純粋SSGならランタイムCPU消費ゼロで、Cloudflare Freeプランの10ms制限に抵触しない
- ISR再生成時のReactレンダリングがCPU制限に抵触するリスクを回避
- クライアントサイドフェッチより高速（追加APIラウンドトリップが不要）

**対応ファイル**:

| ファイル | 変更内容 |
|---------|---------|
| `apps/frontend/src/app/page.tsx` | `revalidate`を設定せず、純粋SSGを維持（変更なし） |

---

### 2.3 #47: ユーザープロフィールページへのISR導入

**課題**: ユーザープロフィールページが毎回サーバーサイドでデータ取得していた

**対応ファイル**:

| ファイル | 変更内容 |
|---------|---------|
| `apps/frontend/src/app/[username]/page.tsx` | `cache: "no-store"` を削除し、ISR `revalidate = 600`（10分）を導入 |

**変更内容**:
```typescript
// ISR: 10分ごとに再生成
export const revalidate = 600;
```

**理由**:
- ユーザープロフィール情報は頻繁に変更されないが、完全に静的ではない
- 10分間隔でのISR再生成が、鮮度とパフォーマンスのバランスとして適切
- R2キャッシュ（#45）と組み合わせることでキャッシュの永続性も確保

---

### 2.4 #48: サイドバーのモーダルコンポーネント遅延ロード

**課題**: TagModalとLoginModalが初期バンドルに含まれ、バンドルサイズが不必要に大きい

**対応ファイル**:

| ファイル | 変更内容 |
|---------|---------|
| `apps/frontend/src/components/Sidebar.tsx` | `TagModal`と`LoginModal`を`next/dynamic`で遅延ロード化 |

**変更内容**:
```typescript
import dynamic from "next/dynamic";

const TagModal = dynamic(
  () => import("@/components/TagModal").then((mod) => mod.TagModal),
  { ssr: false },
);
const LoginModal = dynamic(
  () => import("@/components/LoginModal").then((mod) => mod.LoginModal),
  { ssr: false },
);
```

**理由**:
- モーダルはユーザーがボタンをクリックするまで表示されない
- `ssr: false` でサーバーサイドレンダリングも不要
- 初期ロード時のJavaScriptバンドルサイズを削減し、First Contentful Paintを改善

---

### 2.5 #49: タグ詳細・映画詳細ページへのISR導入

**課題**: 動的ページが毎リクエストでサーバーサイドレンダリングされていた

**対応ファイル**:

| ファイル | 変更内容 | revalidate |
|---------|---------|-----------|
| `apps/frontend/src/app/tags/[tagId]/page.tsx` | ISR導入 | 300秒（5分） |
| `apps/frontend/src/app/movies/[movieId]/page.tsx` | ISR導入 | 3600秒（1時間） |

**revalidate値の根拠**:
- **タグ詳細（5分）**: タグへの映画追加・削除が比較的頻繁に行われるため、短めの間隔
- **映画詳細（1時間）**: 映画情報（TMDB由来）は変更頻度が低いため、長めの間隔

---

## 3. アーキテクチャ上の判断

### Cloudflare Freeプランの制約

Cloudflare Workers Freeプランでは1リクエストあたり10ms CPUの制限がある。
ISR再生成時にはReactのサーバーサイドレンダリングが発生し、CPU時間を消費する。

**対策**:
- ホームページは純粋SSGとし、ランタイムCPU消費をゼロに
- 動的ページ（タグ・映画・ユーザー）はISRで再生成するが、再生成頻度を抑制
- R2キャッシュにより、キャッシュヒット時はCPU消費を最小化

### ISR再生成の仕組み（stale-while-revalidate）

1. `revalidate`期間内のアクセス → キャッシュから即座に返却
2. `revalidate`期間経過後の最初のアクセス → 古いキャッシュを返却しつつ、バックグラウンドで再生成
3. 次回以降のアクセス → 再生成されたキャッシュから返却

---

## 4. デプロイ手順

1. R2バケットの作成（初回のみ）:
   ```bash
   wrangler r2 bucket create cache
   ```

2. 通常のデプロイフロー:
   ```bash
   npm run build
   wrangler deploy
   ```

---

## 5. 変更ファイル一覧

| ファイル | イシュー | 変更種別 |
|---------|---------|---------|
| `apps/frontend/wrangler.jsonc` | #45 | R2バケット設定の有効化 |
| `apps/frontend/open-next.config.ts` | #45 | R2インクリメンタルキャッシュ設定 |
| `apps/frontend/src/components/Sidebar.tsx` | #48 | モーダルの`next/dynamic`遅延ロード |
| `apps/frontend/src/app/[username]/page.tsx` | #47 | ISR `revalidate = 600` 導入 |
| `apps/frontend/src/app/tags/[tagId]/page.tsx` | #49 | ISR `revalidate = 300` 導入 |
| `apps/frontend/src/app/movies/[movieId]/page.tsx` | #49 | ISR `revalidate = 3600` 導入 |
