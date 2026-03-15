# GoSplit

A REST API for splitting expenses between groups — a Splitwise clone built as a learning project.

## Stack

- **Language:** Go (standard library only — `net/http`, `database/sql`)
- **Database:** PostgreSQL
- **Auth:** JWT (`golang-jwt/jwt`)
- **Other:** `lib/pq` driver, `google/uuid`, `godotenv`
- **Infra:** Docker, Docker Compose

No Gin. No GORM.

## Features

- User registration and login with JWT authentication
- Current user profile (`GET /api/users/me`, `PUT /api/users/me`)
- Create and manage groups
- Add and remove group members
- Record expenses with per-user splits
- Update expenses
- Record settlements between users
- Calculate net balances per user in a group
- Swagger docs (`/swagger/`)

## Project Structure

```
cmd/
  main.go                  # Entry point
internal/
  auth/
    handler.go             # POST /api/auth/register, POST /api/auth/login
    service.go
    service_test.go        # TestRegister, TestLogin, TestLogin_WrongPassword, TestLogin_UserNotFound
  groups/
    handler.go             # CRUD + member management
    service.go
    service_test.go        # TestCreateGroup, TestGetGroups, TestGetGroup, TestUpdateGroup, TestDeleteGroup, TestAddMember, TestRemoveMember
  expenses/
    handler.go             # CRUD + splits
    service.go
    service_test.go        # TestCreateExpense, TestGetExpenses, TestGetExpense, TestDeleteExpense
  settlements/
    handler.go             # Create + list settlements
    service.go
    service_test.go        # TestCreateSettlement, TestGetSettlements
  balances/
    handler.go             # GET /api/groups/{id}/balances
    service.go
    service_test.go        # TestGetBalances
  users/
    handler.go             # GET /api/users/me, PUT /api/users/me
    service.go
    service_test.go
migrations/
  001_init.sql             # All 6 tables
pkg/
  database/
    postgres.go            # DB connection
  middleware/
    auth.go                # JWT middleware + GetUserID helper
  models/
    models.go              # Shared structs
```

## Getting Started

### Prerequisites

- Docker + Docker Compose

### Run

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

### Environment Variables

Create a `.env` file in the project root:

```env
POSTGRES_USER=gosplit
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=gosplit_db
POSTGRES_PORT=5432
APP_PORT=8080
JWT_SECRET=your_jwt_secret
DATABASE_URL=postgres://gosplit:yourpassword@localhost:5432/gosplit_db?sslmode=disable
```

## API Reference

### Auth

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| POST | `/api/auth/register` | Register a new user | ❌ |
| POST | `/api/auth/login` | Login and get JWT token | ❌ |

### Groups

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| POST | `/api/groups` | Create a group | ✅ |
| GET | `/api/groups` | List user's groups | ✅ |
| GET | `/api/groups/{id}` | Get a group | ✅ |
| PUT | `/api/groups/{id}` | Update a group | ✅ |
| DELETE | `/api/groups/{id}` | Delete a group | ✅ |
| POST | `/api/groups/{id}/members` | Add a member | ✅ |
| DELETE | `/api/groups/{id}/members/{user_id}` | Remove a member | ✅ |

### Expenses

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| POST | `/api/groups/{id}/expenses` | Create an expense | ✅ |
| GET | `/api/groups/{id}/expenses` | List expenses in a group | ✅ |
| GET | `/api/groups/{id}/expenses/{expenseId}` | Get an expense | ✅ |
| PUT | `/api/groups/{id}/expenses/{expenseId}` | Update an expense | ✅ |
| DELETE | `/api/groups/{id}/expenses/{expenseId}` | Delete an expense | ✅ |

### Settlements

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| POST | `/api/groups/{id}/settlements` | Record a settlement | ✅ |
| GET | `/api/groups/{id}/settlements` | List settlements in a group | ✅ |

### Balances

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| GET | `/api/groups/{id}/balances` | Get net balances for all users in a group | ✅ |

### Users

| Method | Route | Description | Auth |
|--------|-------|-------------|------|
| GET | `/api/users/me` | Get current user profile | ✅ |
| PUT | `/api/users/me` | Update current user profile | ✅ |

## Balance Calculation

A user's balance in a group is calculated as:

```
balance = expenses paid by user
        - splits assigned to user
        + settlements received
        - settlements paid
```

A **positive** balance means the user is owed money.
A **negative** balance means the user owes money.

## Testing

Tests run against real PostgreSQL instances using `testcontainers-go`. Each package spins up an isolated Postgres container, runs the migrations, executes the tests, and tears the container down automatically.

### Prerequisites

- Docker (required for testcontainers)

### Run tests

Run all tests with:

```bash
go test ./internal/...
```

You can also run tests for a single package, e.g.:

```bash
go test ./internal/groups -v
```

The test suites cover the service layer behaviour for `auth`, `groups`, `expenses`, `settlements`, `users` and `balances`.
