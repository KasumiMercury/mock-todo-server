# Mock TODO Server

Go言語ベースの洗練されたモックTODOサーバー。タスク管理のためのREST APIエンドポイントと包括的な認証サポートを提供する。開発環境でOAuth2/OIDCクライアントや様々な認証シナリオをテストするために設計されている。

## クイックスタート

### 基本的な使用方法

デフォルト設定でサーバーを起動：
```bash
./mock-todo-server serve
```

デフォルトでは、ポート8080でJWT認証を有効にして起動

### インタラクティブモード

ガイド付きの設定にはインタラクティブモードを使用：
```bash
./mock-todo-server
```

### 基本的な操作

サーバーの停止:
```bash
./mock-todo-server stop
```

現在のメモリ状態をJSONファイルにエクスポート：
```bash
./mock-todo-server export --memory backup.json
```

ファイルストレージ用のテンプレート出力
```bash
./mock-todo-server export --template
```

### データ永続化

デフォルトではサーバーを停止するとデータは失われる
データを永続化するには、ファイルストレージを使用
```bash
./mock-todo-server serve -f data.json
```

### コマンドラインオプション

#### サーバーコマンド

```bash
# ポートを指定してサーバーを起動
./mock-todo-server serve -p 3000

# 認証なしでサーバーを起動
./mock-todo-server serve -a=false

# セッションベース認証でサーバーを起動
./mock-todo-server serve --auth-mode session

# RSA JWT署名でサーバーを起動
./mock-todo-server serve --jwt-key-mode rsa

# OIDC認証でサーバーを起動
./mock-todo-server serve --auth-mode oidc --oidc-config-path oidc-config.json
```

#### データエクスポートコマンド

```bash
# ファイルベースストレージ用のJSONテンプレートをエクスポート
./mock-todo-server export --template

# テンプレートをカスタムファイルにエクスポート
./mock-todo-server export --template custom.json

# 現在のサーバーメモリ状態をエクスポート
./mock-todo-server export --memory

# メモリ状態をカスタムファイルにエクスポート
./mock-todo-server export --memory backup.json

# OIDC設定テンプレートをエクスポート
./mock-todo-server export --oidc-config

# OIDC設定テンプレートをカスタムファイルにエクスポート
./mock-todo-server export --oidc-config my-oidc-config.json
```

## API ドキュメント

### 認証エンドポイント

#### 標準認証（JWT/セッションモード）

| メソッド | エンドポイント | 説明 |
|--------|-------------|-----|
| POST | `/auth/login` | ユーザーログイン |
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/logout` | ユーザーログアウト |
| GET | `/auth/me` | 現在のユーザー情報を取得 |
| GET | `/auth/jwks` | JSON Web Key Setを取得 |

#### OIDCプロバイダーエンドポイント（OIDCモード）

| メソッド | エンドポイント | 説明 |
|--------|-------------|-----|
| GET/POST | `/auth/authorize` | 認可エンドポイント（ログインフォーム） |
| POST | `/auth/token` | トークンエンドポイント |
| GET | `/auth/userinfo` | ユーザー情報エンドポイント |
| GET | `/auth/jwks` | JSON Web Key Setを取得 |

#### Well-Knownエンドポイント

| メソッド | エンドポイント | 説明 |
|--------|-------------|-----|
| GET | `/.well-known/jwks.json` | 標準JWKSエンドポイント |
| GET | `/.well-known/openid_configuration` | OpenID Connect discovery |

### タスクエンドポイント

| メソッド | エンドポイント | 説明 |
|--------|-------------|-----|
| GET | `/tasks` | 全タスクを取得（認証有効時はユーザー別にフィルター） |
| POST | `/tasks` | 新しいタスクを作成 |
| GET | `/tasks/{id}` | IDでタスクを取得 |
| PUT | `/tasks/{id}` | タスクを更新 |
| DELETE | `/tasks/{id}` | タスクを削除 |

### API使用例

#### 新しいユーザーを登録：
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}'
```

#### ログイン：
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}'
```

#### タスクを作成：
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"プロジェクトドキュメンテーションを完成させる"}'
```

#### 全タスクを取得：
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### タスクを更新：
```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"更新されたタスクタイトル"}'
```

#### タスクを削除：
```bash
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 設定

### 認証モード

1. **JWTモード** (`--auth-mode jwt`): JSON Web Tokensを使用
   - HMAC署名（デフォルト）: `--jwt-key-mode secret`
   - RSA署名: `--jwt-key-mode rsa`

2. **セッションモード** (`--auth-mode session`): Cookieを使用したサーバーサイドセッション

3. **両方モード** (`--auth-mode both`): JWTまたはセッション認証のどちらでも受け入れ

4. **OIDCモード** (`--auth-mode oidc`): OpenID Connectプロバイダーとして動作
   - **設定ファイル必須**: OIDCモードにはJSON設定ファイルが必要
   - クライアントアプリケーションのテスト用OAuth2/OIDCエンドポイントを提供
   - OIDC認証後はJWTトークンを使用してAPIアクセス

#### OIDC設定のセットアップ

OIDCモードでは `--oidc-config-path` で指定する設定ファイルが必要

**設定テンプレートの生成:**
```bash
# OIDC設定テンプレートを生成
./mock-todo-server export --oidc-config oidc-config.json
```

**OIDCモードでサーバー起動:**
```bash
# OIDCモードでサーバーを起動（設定ファイルは必須）
./mock-todo-server serve --auth-mode oidc --oidc-config-path oidc-config.json
```

**設定ファイルの構造:**

OIDC設定ファイルには以下の必須フィールドが含まれている必要がある：

| フィールド | 型 | 必須 | 説明 |
|-----------|----|----- |------|
| `client_id` | string | はい | OAuth2クライアント識別子 |
| `client_secret` | string | はい | OAuth2クライアントシークレット |
| `redirect_uris` | array | はい | 認可コードフロー用の許可されたリダイレクトURI |
| `issuer` | string | はい | OIDC発行者識別子（通常はサーバーURL） |
| `scopes` | array | いいえ | サポートされるスコープ（デフォルト: ["openid", "profile"]） |

**設定例:**
```json
{
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "redirect_uris": [
    "http://localhost:3000/callback",
    "https://your-app.example.com/callback"
  ],
  "issuer": "http://localhost:8080",
  "scopes": [
    "openid",
    "profile"
  ]
}
```

**フィールドの説明:**

- **client_id**: OAuth2クライアントアプリケーションの一意識別子
- **client_secret**: クライアント認証用の秘密キー（安全に保管すること）
- **redirect_uris**: 認証後にユーザーをリダイレクトできる有効なURLの配列
- **issuer**: OIDCプロバイダー（このサーバー）のベースURL
- **scopes**: アプリケーションが要求できる情報スコープのリスト（OIDCにはopenidが必要）

**OIDCフローの例:**

1. **認可リクエスト**: 適切なパラメータで`/auth/authorize`にユーザーを誘導する
2. **ユーザーログイン**: ユーザーがWebフォームで認証する
3. **認可コード**: サーバーが認可コードでリダイレクトする
4. **トークン交換**: `/auth/token`で認可コードをアクセス/IDトークンと交換する
5. **API アクセス**: アクセストークンを使用して保護されたエンドポイントを呼び出す

```bash
# 認可URLの例
http://localhost:8080/auth/authorize?client_id=your-client-id&redirect_uri=http://localhost:3000/callback&response_type=code&scope=openid%20profile

# 認可コードをトークンと交換
curl -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&redirect_uri=http://localhost:3000/callback&client_id=your-client-id&client_secret=your-client-secret"

# アクセストークンを使用してAPI 呼び出し
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### ストレージオプション

1. **メモリストレージ**（デフォルト）: データはメモリに保存され、サーバー停止時に失われる
2. **ファイルストレージ**: データはJSONファイルに永続化される
   ```bash
   ./mock-todo-server serve -f data.json
   ```

### ファイル形式

ファイルストレージを使用する場合のJSON形式：
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "サンプルタスク",
      "user_id": 1,
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "users": [
    {
      "id": 1,
      "username": "user1",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

## ユースケース

### 開発とテスト

- **フロントエンド開発**: React/Vue/Angularアプリケーションのモックバックエンド
- **OAuth2/OIDCテスト**: OAuth2およびOpenID Connectクライアント実装のテスト
- **APIテスト**: 自動テストおよびCI/CDパイプライン用の信頼性の高いエンドポイント
- **認証テスト**: さまざまな認証フローとシナリオのテスト
- **モバイルアプリ開発**: モバイルアプリケーション開発用のモックAPI

### デモとプロトタイピング

- **APIデモンストレーション**: REST APIパターンと認証フローの紹介
- **教育目的**: 動作する実装でOAuth2/OIDCの概念を学習する
- **ラピッドプロトタイピング**: PoC（概念実証）プロジェクトのための迅速なバックエンドセットアップ
