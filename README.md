# Mock TODO Server

A Go-based mock TODO server that provides REST API endpoints for task management with comprehensive authentication support. Designed for testing OAuth2/OIDC clients and various authentication scenarios in development environments.

## Quick Start

### Basic Usage

Start the server with default settings:
```bash
./mock-todo-server serve
```

By default, it starts with JWT authentication enabled on port 8080.

### Interactive Mode

Use interactive mode for guided configuration:
```bash
./mock-todo-server
```

This launches an interactive CLI that guides you through server configuration with helpful prompts and explanations.

### Basic Operations

Stop the server:
```bash
./mock-todo-server stop
```

Export the current memory state to a JSON file:
```bash
./mock-todo-server export memory backup.json
```

Output a template for file storage:
```bash
./mock-todo-server export store
```

The template includes 2 sample users with hashed passwords:
- **user1** (password: `password1`)
- **user2** (password: `password2`)

### Data Persistence

By default, data is stored in memory and lost when the server stops.
To persist data, use file storage:
```bash
./mock-todo-server serve -f data.json
```

### Command-Line Options

#### Server Commands

```bash
# Start the server on a specific port
./mock-todo-server serve -p 3000

# Start the server without authentication
./mock-todo-server serve -a=false

# Start the server with session-based authentication
./mock-todo-server serve --auth-mode session

# Start the server with RSA JWT signing
./mock-todo-server serve --jwt-key-mode rsa

# Start the server with OIDC authentication
./mock-todo-server serve --auth-mode oidc --oidc-config-path oidc-config.json
```

#### Data Export Commands

```bash
# Export a JSON template for file-based storage
./mock-todo-server export store

# Export the template to a custom file
./mock-todo-server export store custom.json

# Template includes sample users:
# - user1 (password: password1)
# - user2 (password: password2)

# Export the current server memory state
./mock-todo-server export memory

# Export the memory state to a custom file
./mock-todo-server export memory backup.json

# Export OIDC configuration template
./mock-todo-server export oidc

# Export OIDC config template to a custom file
./mock-todo-server export oidc my-oidc-config.json
```

## API Documentation

### Authentication Endpoints

#### Standard Authentication (JWT/Session modes)

| Method | Endpoint | Description |
|--------|-------------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/register`| User registration |
| POST | `/auth/logout` | User logout |
| GET | `/auth/me` | Get current user info |
| GET | `/auth/jwks` | Get JSON Web Key Set |

#### OIDC Provider Endpoints (OIDC mode)

| Method | Endpoint | Description |
|--------|-------------|-------------|
| GET/POST | `/auth/authorize` | Authorization endpoint (login form) |
| POST | `/auth/token` | Token endpoint |
| GET | `/auth/userinfo` | User info endpoint |
| GET | `/auth/jwks` | Get JSON Web Key Set |
| GET/POST | `/auth/register` | User registration (web form) |

#### Well-Known Endpoints

| Method | Endpoint | Description |
|--------|-------------|-------------|
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

4. **OIDC Mode** (`--auth-mode oidc`): Acts as an OpenID Connect provider.
   - **Configuration Required**: OIDC mode requires a JSON configuration file
   - Provides OAuth2/OIDC endpoints for testing client applications
   - Uses JWT tokens for API access after OIDC authentication

#### OIDC Configuration Setup

OIDC mode requires a configuration file specified with `--oidc-config-path`. This file defines the OIDC provider settings.

**Generate Configuration Template:**
```bash
# Generate OIDC configuration template
./mock-todo-server export oidc oidc-config.json
```

**Start Server with OIDC:**
```bash
# Start server in OIDC mode (configuration file is mandatory)
./mock-todo-server serve --auth-mode oidc --oidc-config-path oidc-config.json
```

**Configuration File Structure:**

The OIDC configuration file must contain the following required fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `client_id` | string | Yes | OAuth2 client identifier |
| `client_secret` | string | Yes | OAuth2 client secret |
| `redirect_uris` | array | Yes | Allowed redirect URIs for authorization code flow |
| `issuer` | string | Yes | OIDC issuer identifier (typically server URL) |
| `scopes` | array | Optional | Supported scopes (defaults to ["openid", "profile"]) |

**Example Configuration:**
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

**Field Descriptions:**

- **client_id**: Unique identifier for your OAuth2 client application
- **client_secret**: Secret key for client authentication (keep this secure)
- **redirect_uris**: Array of valid URLs where users can be redirected after authentication
- **issuer**: The base URL of your OIDC provider (this server)
- **scopes**: List of information scopes your application can request (openid is required for OIDC)

**User Registration in OIDC Mode:**

OIDC mode includes a web-based user registration system that operates independently of the OIDC authentication flow:

- **Registration URL**: Access `/auth/register` directly (no OIDC parameters required)
- **Independent from OIDC Flow**: Registration is a server-side feature, not part of the OAuth2/OIDC specification
- **Web Interface**: Provides an HTML form for username/password registration
- **Development Focus**: Designed to simplify user creation during frontend development and testing

```bash
# Access registration page directly
http://localhost:8080/auth/register

# Or programmatically register users (in OIDC mode)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=testuser&password=testpass"
```

After registration, users can authenticate through the standard OIDC authorization flow.

**OIDC Flow Example:**

1. **User Registration** (optional): Create test users via `/auth/register`
2. **Authorization Request**: Direct users to `/auth/authorize` with appropriate parameters
3. **User Login**: Users authenticate via the web form
4. **Authorization Code**: Server redirects back with authorization code
5. **Token Exchange**: Exchange code for access/ID tokens at `/auth/token`
6. **API Access**: Use access token to call protected endpoints

```bash
# Example authorization URL
http://localhost:8080/auth/authorize?client_id=your-client-id&redirect_uri=http://localhost:3000/callback&response_type=code&scope=openid%20profile

# Exchange code for tokens
curl -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&redirect_uri=http://localhost:3000/callback&client_id=your-client-id&client_secret=your-client-secret"

# Use access token for API calls
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

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
      "hashed_password": "$2a$10$...",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

**Note**: When using the template export (`export store`), sample users are included with pre-hashed passwords:
- **user1** with password: `password1`
- **user2** with password: `password2`

## Use Cases

### Development and Testing

- **Frontend Development**: Mock backend for React/Vue/Angular applications
- **OAuth2/OIDC Testing**: Test OAuth2 and OpenID Connect client implementations
- **API Testing**: Reliable endpoint for automated testing and CI/CD pipelines
- **Authentication Testing**: Test different authentication flows and scenarios
- **Mobile App Development**: Mock API for mobile application development

### Demo and Prototyping

- **API Demonstrations**: Showcase REST API patterns and authentication flows
- **Educational Purposes**: Learn OAuth2/OIDC concepts with a working implementation
- **Rapid Prototyping**: Quick backend setup for proof-of-concept projects
