# 映画ページからタグ追加ショートカット機能 設計書

**作成日**: 2026-04-12
**ステータス**: 設計中

---

## 1. 概要

映画詳細ページ (`/movies/[movieId]`) から、表示中の映画を **自分が作成した既存タグに追加** または **新規タグを作成して追加** できるショートカット機能を実装する。

### 1.1 背景

- 現在の映画詳細ページには「この映画が含まれるタグ」の閲覧のみで、追加操作の導線がない
- タグに映画を追加するにはタグ詳細ページへ遷移する必要があり、映画起点の操作フローが欠けている
- バックログ（`docs/plans/20260406_cinetag-improvement-backlog.md`）の「映画詳細からの導線強化」に該当

### 1.2 ゴール

- 映画ページから **画面遷移なし** でタグへの映画追加を完結できる
- 既存タグ選択と新規タグ作成を **1つのモーダル内** で提供する

---

## 2. 現状分析

### 2.1 利用可能な既存資産

| 種類 | ファイル | 状態 |
|------|---------|------|
| API関数 | `lib/api/tags/addMovie.ts` (`addMoviesToTag`) | 実装済み |
| API関数 | `lib/api/tags/create.ts` (`createTag`) | 実装済み |
| API関数 | `lib/api/users/listUserTags.ts` (`listUserTags`) | 実装済み |
| API関数 | `lib/api/movies/tags.ts` (`getMovieRelatedTags`) | 実装済み |
| バリデーション | `lib/validation/tag.api.ts` (`TagListItemSchema` 等) | 実装済み |
| バリデーション | `lib/validation/tag.form.ts` (`TagCreateInputSchema`) | 実装済み |
| 共有UI | `components/Modal.tsx` | 実装済み |
| 参考実装 | `components/TagModal.tsx` (タグ作成/編集モーダル) | 実装済み |

### 2.2 未実装（本機能で新規作成が必要）

| 種類 | 内容 |
|------|------|
| コンポーネント | 映画ページ用「タグに追加」ボタン |
| コンポーネント | タグ追加モーダル（既存タグ選択 + 新規作成） |

---

## 3. UI設計

### 3.1 エントリポイント: 「タグに追加」ボタン

映画情報カード内、ジャンルタグの下・あらすじの上に配置する。

```
┌──────────────────────────────────────────────┐
│  [ポスター]   タイトル                        │
│               原題                            │
│               ★ 8.5                          │
│               2024年 / 148分 / アメリカ       │
│               監督: ...                       │
│               [SF] [アクション]               │
│                                               │
│               [ + タグに追加 ]  ← ここ        │
│                                               │
│               あらすじ ...                     │
│               キャスト ...                     │
└──────────────────────────────────────────────┘
```

| 項目 | 仕様 |
|------|------|
| アイコン | `Plus` (lucide-react) |
| ラベル | 「タグに追加」 |
| スタイル | アウトラインボタン。アクセントカラー `#FF5C5C` の border + text |
| 未ログイン時 | ボタンは表示するが、クリック時にサインインページへリダイレクト |
| ログイン済み | クリックでタグ追加モーダルを開く |

### 3.2 タグ追加モーダル

ボタンクリックで開くモーダル。**既存タグへの追加** と **新規タグ作成** を1画面で提供する。

```
┌─────────────────────────────────────────────┐
│  「インセプション」をタグに追加         [×]  │
│─────────────────────────────────────────────│
│                                             │
│  ┌─ あなたのタグから選択 ─────────────────┐ │
│  │  🔍 タグを検索...                      │ │
│  │                                        │ │
│  │  ☐ Mind-Bending Sci-Fi    12作品       │ │
│  │  ☑ 泣ける映画ベスト       8作品  追加済│ │
│  │  ☐ 2024年ベスト映画       5作品        │ │
│  │  ☐ 友達におすすめ         3作品        │ │
│  │  ...                                   │ │
│  └────────────────────────────────────────┘ │
│                                             │
│  ── または ──                               │
│                                             │
│  ┌─ 新しいタグを作成して追加 ─────────────┐ │
│  │  タグ名: [                       ]     │ │
│  │  [ + 作成して追加 ]                    │ │
│  └────────────────────────────────────────┘ │
│                                             │
│                    [ 選択したタグに追加 ]    │
└─────────────────────────────────────────────┘
```

### 3.3 既存タグ選択セクション

**検索対象は自分が作成したタグのみ**とする。他ユーザーのタグは表示しない。

| 項目 | 仕様 |
|------|------|
| データソース | `listUserTags` でログインユーザー自身の `displayId` を指定して取得 |
| 認証 | トークン付きで呼び出し、非公開タグも含めて取得 |
| 検索 | テキスト入力によるクライアントサイドフィルタリング（タグ名の部分一致） |
| 選択方式 | チェックボックスで複数タグ同時選択を許可 |
| 追加済み判定 | `getMovieRelatedTags` の結果と突合し、既にこの映画を含むタグは「追加済み」バッジ + disabled |
| 表示項目 | タグ名、作品数 |
| スクロール | タグ一覧エリアは `max-h-[240px]` で縦スクロール |
| 空状態 | タグ未作成時:「まだタグがありません。下のフォームから作成しましょう」 |
| 検索ヒットなし | 「一致するタグが見つかりません」 |

### 3.4 新規タグ作成セクション（インライン簡易フォーム）

| 項目 | 仕様 |
|------|------|
| 入力フィールド | タグ名のみ（摩擦を最小化） |
| ボタンラベル | 「作成して追加」 |
| バリデーション | `TagCreateInputSchema` を使用（タイトル必須、文字数制限） |
| デフォルト設定 | `is_public: true`, `add_movie_policy: "everyone"` |
| 動作 | `createTag` → 成功後に自動で `addMoviesToTag` → タグ一覧を再取得して反映 |
| 補足テキスト | 「詳細設定はタグページから変更できます」を小さく表示 |

### 3.5 確定ボタン

| 項目 | 仕様 |
|------|------|
| ラベル | 選択数に応じて動的変更: 「タグに追加」→「2件のタグに追加」 |
| 無効状態 | チェックされたタグが0件の場合は disabled |
| 動作 | 選択された各タグに対して `addMoviesToTag` を実行 |
| 完了後 | 成功メッセージ表示 → `movieRelatedTags` クエリを invalidate → モーダルを閉じる |

### 3.6 デザイントーン

既存の `TagModal` に合わせる:

| 項目 | 値 |
|------|-----|
| 背景色 | `#FFF9F3` |
| アクセントカラー | `#FF5C5C` |
| ボーダー | `#F3E1D6` / `#E4D3C7` |
| テキスト（主） | `#1F1A2B` |
| テキスト（副） | `#7C7288` |
| 入力フィールド背景 | `#FFFDF8` |
| 角丸 | `rounded-2xl` / `rounded-3xl` |
| モーダル最大幅 | `max-w-lg` |
| モバイル | フルスクリーンに近いレイアウト（`px-4` でパディング調整） |

---

## 4. 状態遷移

```
[「タグに追加」ボタンクリック]
    │
    ├─ 未ログイン → サインインページへリダイレクト
    │
    └─ ログイン済み → モーダルを開く
         │
         ├─ 既存タグを選択 → 「追加する」ボタンが活性化
         │     └─ クリック → addMoviesToTag × N
         │           ├─ 全成功 → 成功メッセージ → モーダル閉じる → タグ一覧更新
         │           └─ 一部失敗 → エラーメッセージ表示（成功分は反映）
         │
         └─ 新規タグ名を入力 → 「作成して追加」ボタンが活性化
               └─ クリック → createTag → addMoviesToTag
                     ├─ 成功 → タグ一覧に追加済みとして反映（モーダルは閉じない）
                     └─ 失敗 → エラーメッセージ表示
```

---

## 5. コンポーネント設計

### 5.1 構成

```
MovieDetailClient.tsx
  └─ AddToTagButton                (トリガーボタン)
      └─ AddToTagModal             (モーダル本体)
           ├─ UserTagCheckList     (自分のタグ一覧 + 検索 + チェックボックス)
           └─ QuickCreateTagForm   (インライン新規作成フォーム)
```

### 5.2 各コンポーネントの責務

| コンポーネント | 配置先 | 責務 |
|---|---|---|
| `AddToTagButton` | `app/movies/[movieId]/_components/` | ログイン判定、モーダル開閉の制御 |
| `AddToTagModal` | `app/movies/[movieId]/_components/` | モーダル全体のレイアウト、追加実行ロジック、状態管理 |
| `UserTagCheckList` | `AddToTagModal` 内部 | `listUserTags` 取得、検索フィルタ、チェック状態、追加済み判定 |
| `QuickCreateTagForm` | `AddToTagModal` 内部 | タグ名入力、`createTag` + `addMoviesToTag` の実行 |

### 5.3 Props設計

```typescript
// AddToTagButton
type AddToTagButtonProps = {
  tmdbMovieId: number;
  movieTitle: string;
  relatedTagIds: string[];  // 既に映画が含まれるタグIDの配列
};

// AddToTagModal
type AddToTagModalProps = {
  open: boolean;
  onClose: () => void;
  tmdbMovieId: number;
  movieTitle: string;
  relatedTagIds: string[];
};
```

---

## 6. データフロー

### 6.1 モーダル表示時のデータ取得

```
1. useAuth() → ユーザーの displayId を取得
2. listUserTags({ displayId, token }) → 自分のタグ一覧を取得
3. relatedTagIds (親から受け取り) → 追加済みタグの判定に使用
```

### 6.2 既存タグへの追加

```
1. ユーザーがチェックボックスで1つ以上のタグを選択
2. 「追加する」ボタンをクリック
3. 選択された各タグに対して addMoviesToTag を実行:
   addMoviesToTag({
     tagId: selectedTagId,
     token: authToken,
     movies: [{ tmdb_movie_id: tmdbMovieId }]
   })
4. 全完了後:
   - queryClient.invalidateQueries(["movieRelatedTags", tmdbMovieId])
   - モーダルを閉じる
```

### 6.3 新規タグ作成 + 追加

```
1. ユーザーがタグ名を入力して「作成して追加」をクリック
2. createTag({ token, input: { title, is_public: true, add_movie_policy: "everyone" } })
3. 成功後、返却された tagId を使って addMoviesToTag を実行
4. 完了後:
   - queryClient.invalidateQueries(["userTags", displayId])  // タグ一覧を更新
   - queryClient.invalidateQueries(["movieRelatedTags", tmdbMovieId])
   - 新規タグをチェック済み + 追加済み状態で一覧に反映
   - モーダルは閉じない（続けて他のタグにも追加可能）
```

---

## 7. 利用するAPI

| API関数 | 用途 | 認証 |
|---------|------|------|
| `listUserTags({ displayId, token })` | ログインユーザーの作成タグ一覧取得 | 必須（非公開タグも含む） |
| `getMovieRelatedTags(tmdbMovieId)` | 映画に紐づくタグ一覧（追加済み判定用、親で取得済み） | 不要 |
| `addMoviesToTag({ tagId, token, movies })` | タグへ映画を追加 | 必須 |
| `createTag({ token, input })` | 新規タグ作成 | 必須 |

---

## 8. React Query キー設計

| クエリキー | 用途 | invalidate タイミング |
|-----------|------|----------------------|
| `["userTags", displayId]` | 自分のタグ一覧 | 新規タグ作成後 |
| `["movieRelatedTags", tmdbMovieId]` | 映画の関連タグ一覧 | タグ追加完了後 |

---

## 9. エラーハンドリング

| シナリオ | 対応 |
|---------|------|
| タグ一覧取得失敗 | モーダル内にエラーメッセージ + リトライボタン |
| 映画追加失敗（権限なし等） | 該当タグ横にエラー表示 |
| タグ作成失敗（バリデーション） | フォーム下にエラーメッセージ |
| タグ作成失敗（サーバーエラー） | フォーム下にエラーメッセージ |
| 既に追加済み（`already_exists`） | エラーとせず正常完了扱い（`addMoviesToTag` が 207 で返す） |

---

## 10. 実装順序

### Phase 1: モーダルコンポーネント作成

1. `AddToTagModal` のUI骨格を作成（`Modal` ラッパー使用）
2. `UserTagCheckList` の実装（`listUserTags` 連携、検索、チェックボックス）
3. `QuickCreateTagForm` の実装（タグ名入力、作成 + 追加ロジック）
4. 確定ボタンの `addMoviesToTag` 連携

### Phase 2: 映画ページへの統合

1. `AddToTagButton` の作成（ログイン判定、モーダル開閉）
2. `MovieDetailClient` への組み込み
3. `relatedTagIds` の受け渡し

### Phase 3: UX改善

1. 追加完了後のクエリ invalidate
2. ローディング・エラー状態の仕上げ
3. モバイル対応の確認・調整

---

## 11. 将来の拡張候補

- **映画検索結果ページ**からの「タグに追加」（`AddToTagModal` を再利用）
- **他ユーザーのタグへの追加**（`add_movie_policy: "everyone"` のタグを検索対象に含める）
- **バッチ追加**（複数映画を一度にタグへ追加）
- **タグ追加時のメモ入力**（`addMoviesToTag` の `note` パラメータを活用）
