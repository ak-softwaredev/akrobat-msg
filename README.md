# Akrobat Msg

Akrobat Msg is a messaging service with a Go HTTP API, PostgreSQL persistence, and per-chat WebSocket events. The API is built with Gin; database migrations are embedded in the server and applied automatically at startup.

The repository also includes a vibecoded React client under `frontend/`. This README focuses on the backend and its API contract. Do not rely on it's interface in production, as the frontend code was not reviewed by a developer, and might be vulnerable to attacks. I do not know frontend.

## Requirements

- Go 1.27.2 or newer
- PostgreSQL 16 for local development, or Docker Compose
- Docker Compose v2 for the containerized setup

## Run With Docker Compose

Compose starts PostgreSQL, the API, and the frontend reverse proxy. Create a project-root `.env` file (Compose requires it for the backend service):

```dotenv
POSTGRES_USER=app_user
POSTGRES_PASSWORD=change-this-password
POSTGRES_DB=app_db
```

Start the services:

```sh
docker compose up --build
```

The API container waits for PostgreSQL's health check, applies migrations, and listens on port `8080` inside the Compose network. The frontend proxy is exposed at `http://localhost:8080`. To stop the services, run `docker compose down`; the PostgreSQL data remains in the `postgres_data` volume. Add `-v` to remove that volume as well.

## Run The Backend Locally

Start PostgreSQL and set the connection string and listen address. `PORT` uses Gin's address format, such as `:8080`.

```sh
export DATABASE_URL='postgresql://app_user:change-this-password@localhost:5432/app_db?sslmode=disable'
export PORT=':8080'
go run ./cmd/server
```

Alternatively, build and run the binary with `make build` and `make run`. The server reads `.env` from its working directory when present. It applies migrations before opening the database pool, then starts the HTTP server. A database connection is required at startup.

## Configuration

| Variable | Description | Default / Compose value |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL connection URL; required by the server | Set in the local environment; Compose constructs it from `POSTGRES_*` values and uses `postgres` as the hostname |
| `PORT` | Gin listen address | `:8080` |
| `GIN_MODE` | Gin runtime mode | Compose sets `release` |
| `POSTGRES_USER` | PostgreSQL username | `app_user` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `app_password` |
| `POSTGRES_DB` | PostgreSQL database name | `app_db` |

The three `POSTGRES_*` defaults are defined in `compose.yaml`. For local runs, ensure the credentials in `DATABASE_URL` match your PostgreSQL instance.

## HTTP API

All request and response bodies are JSON unless the response is `204 No Content`. Authenticated routes accept `Authorization: Bearer <auth_token>`. Chat-specific routes also require the authenticated user to be a member of that chat. Error responses generally use an `error` object with a `message` field.

### Health

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health/live` | Process liveness; does not query PostgreSQL |
| `GET` | `/health/ready` | Readiness check; returns `503` if PostgreSQL cannot be reached |

### Authentication

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/auth/register` | Register and create a session (`201`) |
| `POST` | `/api/v1/auth/login` | Log in and create a session (`200`) |
| `POST` | `/api/v1/auth/logout` | Revoke the token in the Authorization header (`204`) |

Register and login accept `{"username":"...","password":"..."}` and return `id`, `username`, and `auth_token`. Usernames must be 3-20 characters and passwords 8-64 characters. Auth tokens expire after seven days.

Example:

```sh
curl -sS http://localhost:8080/api/v1/auth/register \
	-H 'Content-Type: application/json' \
	-d '{"username":"alex","password":"correct-horse-battery"}'
```

### Users

All user routes require authentication.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/users/me` | Get the authenticated user's profile |
| `GET` | `/api/v1/users/:id` | Get a user by ID |
| `PATCH` | `/api/v1/users/me` | Update one profile field |
| `DELETE` | `/api/v1/users/me` | Delete the authenticated user's account (`204`) |

Profile updates accept `{"field":"bio","value":"..."}`. Supported fields are `username`, `password`, and `bio`.

### Chats

All chat routes require authentication. Routes containing a chat ID require chat membership; administrative operations require the appropriate chat role.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/chats` | List the authenticated user's chats |
| `POST` | `/api/v1/chats` | Create a chat with `{"name":"..."}` (`201`); creator becomes owner |
| `GET` | `/api/v1/chats/:id` | Get chat details and members |
| `PATCH` | `/api/v1/chats/:id` | Update `name` or `description` with `{"field":"...","value":"..."}` |
| `DELETE` | `/api/v1/chats/:id` | Delete a chat (owner only; `204`) |
| `POST` | `/api/v1/chats/:id/members` | Add a member with `{"id":<user_id>}` (owner/admin; `204`) |
| `DELETE` | `/api/v1/chats/:id/members/:rid` | Remove a member by user ID (owner/admin; `204`) |

### Messages

Message routes require authentication and membership in the specified chat.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/chats/:id/messages` | List messages; accepts `beforeID` and `limit` query parameters |
| `POST` | `/api/v1/chats/:id/messages` | Create a message with `{"content":"..."}` (`201`) |
| `PATCH` | `/api/v1/chats/:id/messages/:mid` | Update a message with `{"content":"..."}` |
| `DELETE` | `/api/v1/chats/:id/messages/:mid` | Delete a message (`204`) |

The message list defaults to a limit of 50. `beforeID` requests messages before a message ID for pagination. Message content must be 1-4000 characters. Only the message sender can update or delete that message.

## WebSocket Events

Connect to `/api/v1/chats/:id/ws` to receive events for a chat. The connection requires chat membership and authentication. WebSocket clients may pass the token as a query parameter: `ws://localhost:8080/api/v1/chats/1/ws?token=<auth_token>`.

Events are JSON objects with a `type` and `data` field:

```json
{"type":"message.created","data":{"id":1,"chat_id":1,"sender_id":1,"content":"Hello"}}
```

Supported event types are `message.created`, `message.updated`, and `message.deleted`. The first two contain the message; the delete event's data contains `id` and `chat_id`.

## Development

```sh
make build  # build bin/akrobat-msg
make run    # build and run the server
make test   # run Go tests with the race detector and coverage
make fmt    # format Go files
make vet    # run go vet
make tidy   # tidy Go module dependencies
```

## Backend Structure

- `cmd/server`: application startup, dependency wiring, and route registration
- `internal/handler`: HTTP handlers and request/response binding
- `internal/middleware`: authentication and chat membership checks
- `internal/service`: business rules
- `internal/repository`: PostgreSQL data access and session storage
- `internal/database/migrations`: embedded SQL schema migrations
- `internal/ws`: WebSocket connection handling and event distribution