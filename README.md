# HEIC to JPG Converter

HEICファイルをJPG形式に変換するWebアプリケーション

## 特徴

- **シングル＆バッチ変換**: 1ファイルずつ、または複数ファイルを一括変換
- **ドラッグ&ドロップ**: 直感的なファイルアップロード
- **品質設定**: JPG出力時の画質を1-100で調整可能
- **個別ダウンロード**: 変換後のファイルを個別にダウンロード
- **リアルタイム進捗**: 各ファイルの変換状況をリアルタイム表示
- **クリーンアーキテクチャ**: バックエンドは Go、フロントエンドは React (TypeScript)
- **ロギング統合**: vibe-coding-loggerによる詳細なロギング

## 技術スタック

### バックエンド
- **言語**: Go 1.23+
- **フレームワーク**: Gin
- **アーキテクチャ**: クリーンアーキテクチャ（Domain, Usecase, Infrastructure, Handler）
- **HEIC変換**: github.com/adrium/goheif
- **ロギング**: vibe-coding-logger

### フロントエンド
- **言語**: TypeScript
- **ライブラリ**: React 18
- **ビルドツール**: Vite
- **HTTP クライアント**: Axios
- **テスト**: Vitest, React Testing Library

### インフラ
- **コンテナ**: Docker + Docker Compose
- **開発環境**: ホットリロード対応

## 必須要件

- Docker Desktop
- Git

## インストール

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd heic-to-jpg
```

### 2. 開発環境の起動

```bash
docker-compose up
```

初回起動時は、依存関係のインストールとビルドが行われます。起動完了後、以下のURLでアクセスできます：

- **フロントエンド**: http://localhost:3000
- **バックエンドAPI**: http://localhost:8080
- **ヘルスチェック**: http://localhost:8080/health

## 使い方

### 基本操作

1. **ファイルのアップロード**
   - ブラウザで http://localhost:3000 を開く
   - HEICファイルをドラッグ&ドロップ、またはファイル選択ボタンをクリック
   - 複数ファイルの選択が可能

2. **品質設定**
   - スライダーでJPG品質を調整（1-100、デフォルト: 85）
   - 品質が高いほどファイルサイズも大きくなります

3. **変換実行**
   - 各ファイルの「変換」ボタンをクリック（個別変換）
   - または「すべて変換」ボタンで一括変換

4. **ダウンロード**
   - 変換完了後、「ダウンロード」ボタンで保存
   - ファイル名は自動的に `.heic` → `.jpg` に変更されます

## API仕様

### エンドポイント

#### `POST /api/convert/single`

単一ファイルの変換

**リクエスト:**
- Content-Type: `multipart/form-data`
- Parameters:
  - `file`: HEIC ファイル（必須）
  - `quality`: JPG品質 1-100（オプション、デフォルト: 85）

**レスポンス:**
- Content-Type: `image/jpeg`
- Body: 変換後のJPGファイル

#### `POST /api/convert/batch`

複数ファイルのバッチ変換

**リクエスト:**
- Content-Type: `multipart/form-data`
- Parameters:
  - `files`: HEIC ファイル配列（必須、最大100ファイル）
  - `quality`: JPG品質 1-100（オプション、デフォルト: 85）

**レスポンス:**
```json
{
  "success_count": 2,
  "error_count": 0,
  "results": [
    {
      "file_name": "image1.jpg",
      "original_size": 1024000,
      "converted_size": 512000,
      "data": [...]
    }
  ],
  "errors": {}
}
```

#### `GET /health`

ヘルスチェック

**レスポンス:**
```json
{
  "status": "ok"
}
```

## プロジェクト構成

```
heic-to-jpg/
├── frontend/                # React アプリケーション
│   ├── src/
│   │   ├── components/      # UI コンポーネント
│   │   ├── services/        # API通信、ロギング
│   │   ├── types/           # TypeScript型定義
│   │   └── test/            # テストセットアップ
│   ├── Dockerfile
│   ├── package.json
│   └── vite.config.ts
├── backend/                 # Go API サーバー
│   ├── cmd/server/          # エントリーポイント
│   ├── internal/
│   │   ├── domain/          # エンティティ、ビジネスルール
│   │   ├── usecase/         # アプリケーションロジック
│   │   ├── handler/         # HTTPハンドラー
│   │   └── infrastructure/  # HEIC変換実装
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   └── go.mod
├── vibe-coding-logger/      # ロギングライブラリ
├── docker-compose.yml       # 開発環境定義
└── README.md                # このファイル
```

## 開発

### テストの実行

**バックエンド:**
```bash
cd backend
go test ./... -v
```

**フロントエンド:**
```bash
cd frontend
npm test
```

### ビルド

**バックエンド:**
```bash
cd backend
go build ./cmd/server
```

**フロントエンド:**
```bash
cd frontend
npm run build
```

### 開発ガイドライン

詳細は以下のドキュメントを参照：

- [CLAUDE.md](CLAUDE.md) - 開発方針全般
- [CLAUDE_TDD.md](CLAUDE_TDD.md) - TDDとテスト戦略
- [CLAUDE_GIT.md](CLAUDE_GIT.md) - Git管理・バージョン管理
- [CLAUDE_CICD.md](CLAUDE_CICD.md) - CI/CD設定
- [CLAUDE_SECURITY.md](CLAUDE_SECURITY.md) - セキュリティ・安全性ルール
- [CLAUDE_LOGGER.md](CLAUDE_LOGGER.md) - vibe-coding-logger統合
- [CLAUDE_BACKLOG.md](CLAUDE_BACKLOG.md) - プロダクトバックログ管理

## トラブルシューティング

### Docker コンテナが起動しない

```bash
# コンテナを停止
docker-compose down

# キャッシュをクリアして再起動
docker-compose up --build
```

### ポートがすでに使用されている

docker-compose.yml を編集してポート番号を変更：

```yaml
services:
  frontend:
    ports:
      - "3001:3000"  # 3000 → 3001 に変更
  backend:
    ports:
      - "8081:8080"  # 8080 → 8081 に変更
```

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

---

開発開始日: 2025-12-31
