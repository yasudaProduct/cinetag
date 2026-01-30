# cinetag - 映画タグ共有サービス

## 1. サービス概要

映画に対してユーザーが自由に「タグ（プレイリストのようなテーマ）」を作成し、他ユーザーもタグに映画を登録できる共有サービス。

## 2. 環境構築

### 前提条件

- Go 1.25 以上
- Node.js 18 以上
- Docker と Docker Compose
- ngrok（Webhook 受信用、任意）

### 全体の起動手順

1. **リポジトリのクローン**

   ```bash
   git clone <repository-url>
   cd cinetag
   ```

2. **PostgreSQL の起動**

   プロジェクトルートで Docker Compose を使用して PostgreSQL を起動します。

   ```bash
   docker compose up -d postgres
   ```

3. **バックエンドのセットアップと起動**

   詳細は [apps/backend/README.md](apps/backend/README.md) を参照してください。

4. **フロントエンドのセットアップと起動**

   詳細は [apps/frontend/README.md](apps/frontend/README.md) を参照してください。

## 3. 開発


### Git 運用フロー

#### 1. ブランチ戦略

- **main** : 本番運用用のブランチ。常にデプロイ可能な状態を保つ
- **develop** : 開発用の統合ブランチ
- **feature/**: 機能追加や修正ごとに `feature/xxxx` ブランチを作成
- **fix/**: バグ修正用の `fix/xxxx` ブランチを作成

### git worktree

1. **新しい作業ディレクトリを追加する**

   ```bash
   # 例: feature/awesome を apps/awesome-feature ディレクトリで作業したい場合
   git worktree add -b feature/awesome ../apps/awesome-feature origin/feature/awesome
   ```

   - `-b feature/awesome` : 新しいブランチを作成
   - `../apps/awesome-feature` : チェックアウト先ディレクトリ
   - `origin/feature/awesome` : リモートブランチをベースに作成

2. **既存ブランチを別ディレクトリでチェックアウトする**

   ```bash
   git worktree add ../staging develop
   ```

   - `../staging` ディレクトリで `develop` ブランチの内容を直接編集可能

3. **ワークツリーの一覧表示**

   ```bash
   git worktree list
   ```

4. **不要になった作業ディレクトリを削除する**

   ```bash
   git worktree remove ../apps/awesome-feature
   ```

詳細な解説: [Git公式マニュアル（worktree）](https://git-scm.com/docs/git-worktree)





---




### 詳細

- バックエンドの詳細: [apps/backend/README.md](apps/backend/README.md)
- フロントエンドの詳細: [apps/frontend/README.md](apps/frontend/README.md)
- ドキュメント一覧: [docs/README.md](docs/README.md)
- API 仕様: [docs/api/api-spec.md](docs/api/api-spec.md)
- アーキテクチャ（バックエンド）: [docs/architecture/backend-architecture.md](docs/architecture/backend-architecture.md)
- CI/CD: [docs/operations/cicd.md](docs/operations/cicd.md)

