# Mock TODO Server

A mock TODO server that provides REST API endpoints for task management with flexible authentication support.

## Quick Start

### Basic Usage

Start the server with default settings:
```bash
./mock-todo-server serve
```

By default, it starts with JWT authentication enabled on port 8080.

Stop the server:
```bash
./mock-todo-server stop
```

By default, data is lost when the server stops.
To persist data, use file storage:
```bash
./mock-todo-server serve -f data.json
```

Export the current memory state to a JSON file:
```bash
./mock-todo-server export --memory backup.json
```

Output a template for file storage:
```bash
./mock-todo-server export --template
```

### Command-Line Options

#### Server Commands

```bash
# Start the server on a specific port
./mock-todo-server serve -p 3000

# Start the server without authentication
./mock-todo-server serve -a false

# Start the server with session-based authentication
./mock-todo-server serve --auth-mode session

# Start the server with RSA JWT signing
./mock-todo-server serve --jwt-key-mode rsa
```

#### Data Export Commands

```bash
# Export a JSON template for file-based storage
./mock-todo-server export --template

# Export the template to a custom file
./mock-todo-server export --template custom.json

# Export the current server memory state
./mock-todo-server export --memory

# Export the memory state to a custom file
./mock-todo-server export --memory backup.json
```

## API Documentation

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|-------------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/register`| User registration |
| POST | `/auth/logout` | User logout |
| GET | `/auth/me` | Get current user info |
| GET | `/auth/jwks` | Get JSON Web Key Set |
| GET | `/.well-known/jwks.json` | Standard JWKS endpoint |
| GET | `/.well-known/openid_configuration` | OpenID Connect discovery |

### Task Endpoints

| Method | Endpoint | Description |
|--------|-------------|-------------|
| GET | `/tasks` | Get all tasks (filtered by user if auth is enabled) |
| POST | `/tasks` | Create a new task |
| GET | `/tasks/{id}` | Get a task by ID |
| PUT | `/tasks/{id}` | Update a task |
| DELETE | `/tasks/{id}` | Delete a task |

### API Usage Examples

#### Register a new user:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}'
```

#### Log in:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}'
```

#### Create a task:
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"Complete project documentation"}'
```

#### Get all tasks:
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Update a task:
```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"Updated task title"}'
```

#### Delete a task:
```bash
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Configuration

### Authentication Modes

1. **JWT Mode** (`--auth-mode jwt`): Uses JSON Web Tokens.
   - HMAC signature (default): `--jwt-key-mode secret`
   - RSA signature: `--jwt-key-mode rsa`

2. **Session Mode** (`--auth-mode session`): Server-side sessions using HTTP cookies.

3. **Both Mode** (`--auth-mode both`): Accepts either JWT or session authentication.

### Storage Options

1. **Memory Storage** (default): Data is stored in memory and lost on server shutdown.
2. **File Storage**: Data is persisted to a JSON file.
   ```bash
   ./mock-todo-server serve -f data.json
   ```

### File Format

JSON format for file storage:
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Sample Task",
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