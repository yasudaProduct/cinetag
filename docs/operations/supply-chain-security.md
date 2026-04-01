# サプライチェーン攻撃対策

参考記事: [サプライチェーン攻撃の最新動向と対策](https://zenn.dev/dely_jp/articles/supply-chain-kowai)

## 背景

サプライチェーン攻撃の約8割は、パッケージ公開から1週間以内に検出・削除されている。この知見をもとに、以下の5つの対策を実施した。

---

## 実施した対策

### 1. クールダウン期間の設定（min-release-age）

**対象ファイル:** `apps/frontend/.npmrc`

公開から7日未満のパッケージを取り込まないよう設定。公開直後に悪意あるコードが混入するリスクを低減する。

```ini
min-release-age=7
```

**トレードオフ:** 緊急セキュリティパッチの適用が最大7日遅延する可能性がある。ただし、マルウェアがインストールされるリスクと比較して許容できると判断した。

---

### 2. ロックファイルの厳格な管理

**対象ファイル:** 全 GitHub Actions ワークフロー

CI/CD では `npm ci` を使用しており、`package-lock.json` に記録されたバージョンのみをインストールする。これにより、意図しないバージョンアップやトランジティブ依存の変化を防ぐ。

既存のワークフローはすでに `npm ci` を使用していたため、追加の変更は不要。

また `package.json` の `overrides` フィールドで脆弱なトランジティブ依存をピン留めしている（既存設定）。

---

### 3. インストールスクリプトの無効化（ignore-scripts）

**対象ファイル:** `apps/frontend/.npmrc`

`postinstall` などのライフサイクルスクリプトを無効化し、パッケージインストール時に悪意あるコードが実行されるリスクを排除する。

```ini
ignore-scripts=true
```

**注意:** ネイティブアドオン（node-gypなど）を必要とするパッケージはビルドが失敗する場合がある。現在のフロントエンド依存関係はすべてこの設定で問題ないことを確認済み。

---

### 4. GitHub Actions の SHA ピン留め

**対象ファイル:**
- `.github/workflows/ci-pr.yml`
- `.github/workflows/ci-develop.yml`
- `.github/workflows/ci-main.yml`
- `.github/workflows/dependency-review.yml`

タグ参照（例: `actions/checkout@v4`）をコミットハッシュに置き換えた。タグは書き換え可能なため、ハッシュによるピン留めで改ざんを防ぐ。

| アクション | タグ | コミット SHA |
|---|---|---|
| `actions/checkout` | v4 | `34e114876b0b11c390a56381ad16ebd13914f8d5` |
| `actions/setup-go` | v5 | `40f1582b2485089dde7abd97c1529aa768e1baff` |
| `actions/setup-node` | v4 | `49933ea5288caeca8642d1e84afbd3f7d6820020` |
| `google-github-actions/auth` | v2 | `c200f3691d83b41bf9bbd8638997a462592937ed` |
| `google-github-actions/deploy-cloudrun` | v2 | `251330ba9a8a34bfbc1622895f42e1d53fd14522` |
| `actions/dependency-review-action` | v4.9.0 | `2031cfc080254a8a887f58cffee85186f0e49e48` |

各行にはコメントでタグ名を残しており、可読性を維持している。

**SHA の更新方法:** [pinact](https://github.com/suzuki-shunsuke/pinact) を使うと自動で最新 SHA に更新できる。

```bash
# pinact のインストールと実行例
go install github.com/suzuki-shunsuke/pinact/cmd/pinact@latest
pinact run
```

---

### 5. Dependency Review Action の導入

**対象ファイル:** `.github/workflows/dependency-review.yml`

Pull Request 時に依存関係の変更をスキャンし、既知の脆弱性が混入していないかチェックする。`moderate` 以上の重大度でジョブが失敗するよう設定。

```yaml
- uses: actions/dependency-review-action@2031cfc080254a8a887f58cffee85186f0e49e48 # v4.9.0
  with:
    fail-on-severity: moderate
    comment-summary-in-pr: always
```

PRにはスキャン結果がコメントとして自動投稿される。

---

## 対策一覧まとめ

| 対策 | 実施内容 | 対象ファイル |
|---|---|---|
| クールダウン期間 | `min-release-age=7` | `apps/frontend/.npmrc` |
| ロックファイル管理 | `npm ci` 使用（既存） | CI ワークフロー全般 |
| インストールスクリプト無効化 | `ignore-scripts=true` | `apps/frontend/.npmrc` |
| GitHub Actions SHA ピン留め | タグ → コミットハッシュ | `.github/workflows/*.yml` |
| Dependency Review | PR 時に自動脆弱性スキャン | `.github/workflows/dependency-review.yml` |
