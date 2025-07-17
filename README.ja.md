# Mock TODO Server

柔軟な認証サポートを備えたタスク管理用のREST APIエンドポイントを提供する、モックTODOサーバー

## クイックスタート

### 基本的な使用方法

デフォルト設定でサーバーを起動：
```bash
./mock-todo-server serve
```

デフォルトでは、ポート8080でJWT認証を有効にして起動

サーバーの停止:
```bash
./mock-todo-server stop
```

デフォルトではサーバーを停止するとデータは失われる
データを永続化するには、ファイルストレージを使用する
```bash
./mock-todo-server serve -f data.json
```

現在のメモリ状態をJSONファイルにエクスポート：
```bash
./mock-todo-server export --memory backup.json
```

ファイルストレージ用のテンプレート出力
```bash
./mock-todo-server export --template
```

### コマンドラインオプション

#### サーバーコマンド

```bash
# ポートを指定してサーバーを起動
./mock-todo-server serve -p 3000

# 認証なしでサーバーを起動
./mock-todo-server serve -a false

# セッションベース認証でサーバーを起動
./mock-todo-server serve --auth-mode session

# RSA JWT署名でサーバーを起動
./mock-todo-server serve --jwt-key-mode rsa
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
```

## API ドキュメント

### 認証エンドポイント

| メソッド | エンドポイント | 説明 |
|--------|-------------|-----|
| POST | `/auth/login` | ユーザーログイン |
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/logout` | ユーザーログアウト |
| GET | `/auth/me` | 現在のユーザー情報を取得 |
| GET | `/auth/jwks` | JSON Web Key Setを取得 |
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

2. **セッションモード** (`--auth-mode session`): HTTPクッキーを使用したサーバーサイドセッション

3. **両方モード** (`--auth-mode both`): JWTまたはセッション認証のどちらでも受け入れ

### ストレージオプション

1. **メモリストレージ**（デフォルト）: データはメモリに保存され、サーバー停止時に失われます
2. **ファイルストレージ**: データはJSONファイルに永続化されます
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
