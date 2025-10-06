# Task Scheduler API

A RESTful Task Scheduler Service built in Go that manages HTTP tasks with persistent storage and reliable execution tracking. It supports both one-off and recurring (cron) tasks with comprehensive result logging.

## Features

- ✅ **Task Management**: Create, read, update, and cancel HTTP tasks
- ✅ **Flexible Scheduling**: Support for both one-off and cron-based recurring tasks
- ✅ **Persistent Storage**: PostgreSQL database with automatic migrations
- ✅ **Execution Tracking**: Detailed logging of every task execution with results
- ✅ **Docker Support**: Fully containerized with Docker Compose
- ✅ **Graceful Shutdown**: Proper handling of application lifecycle

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15+ (if running without Docker)

### Running with Docker Compose (Recommended)

1. Clone the repository:

```bash
git clone https://github.com/ayushsarode/task-scheduler.git
cd task-scheduler
```

2. Start the services:

```bash
cd deployments
docker-compose up -d
```

3. The API will be available at `http://localhost:8080`

4. Check service health:

```bash
curl http://localhost:8080/health
```

### Running Locally

1. Set up PostgreSQL database and update the `.env` file:

```bash
cp .env.example .env
# Edit .env with your database configuration
```

2. Install dependencies:

```bash
go mod download
```

3. Run database migrations:

```bash
go run cmd/server/main.go
```

4. The application will automatically run migrations on startup.

## API Documentation

### Base URL

```
http://localhost:8080
```

### Endpoints

#### Health Check

- `GET /health` - Service health status

#### Tasks

- `POST /api/v1/tasks` - Create a new task
- `GET /api/v1/tasks` - List all tasks (with pagination and filtering)
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Cancel task
- `GET /api/v1/tasks/{id}/results` - Get task execution results

#### Results

- `GET /api/v1/results` - List all task results (with filtering)

### � **API Examples**

#### Create a One-off Task

Execute a webhook notification in 5 minutes:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Order Confirmation Webhook",
    "trigger": {
      "type": "one-off",
      "datetime": "2025-10-06T20:10:00Z"
    },
    "action": {
      "method": "POST",
      "url": "https://webhook.site/your-unique-endpoint",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer sk-webhook-token-123",
        "X-Source": "task-scheduler"
      },
      "payload": {
        "event": "order.confirmed",
        "order_id": "ORD-2025-001",
        "customer_email": "user@example.com",
        "amount": 99.99,
        "timestamp": "2025-10-06T20:10:00Z"
      }
    }
  }'
```

#### Create a Recurring Task (Cron)

Daily backup status check every morning at 9 AM:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Backup Health Check",
    "trigger": {
      "type": "cron",
      "cron": "0 0 9 * * *"
    },
    "action": {
      "method": "GET",
      "url": "https://api.backupservice.com/v1/status",
      "headers": {
        "Authorization": "Bearer backup-api-key-456",
        "User-Agent": "TaskScheduler/1.0",
        "Accept": "application/json"
      }
    }
  }'
```

#### Frequent Testing Task

For testing - ping every 30 seconds:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Quick Test Ping",
    "trigger": {
      "type": "cron",
      "cron": "*/30 * * * * *"
    },
    "action": {
      "method": "GET",
      "url": "https://httpbin.org/get",
      "headers": {
        "User-Agent": "TaskScheduler-Test/1.0"
      }
    }
  }'
```

#### List and Filter Tasks

```bash
# Get all scheduled tasks
curl "http://localhost:8080/api/v1/tasks?status=scheduled&page=1&limit=10"

# Get completed tasks
curl "http://localhost:8080/api/v1/tasks?status=completed"

# Search tasks by name
curl "http://localhost:8080/api/v1/tasks?name=backup"
```

#### View Task Execution Results

```bash
# Get results for a specific task
curl "http://localhost:8080/api/v1/tasks/{task-id}/results"

# Get successful results only
curl "http://localhost:8080/api/v1/results?success=true&page=1&limit=5"

# Get recent failed executions
curl "http://localhost:8080/api/v1/results?success=false&limit=10"
```

#### Update or Cancel Tasks

```bash
# Update a task (change schedule or action)
curl -X PUT http://localhost:8080/api/v1/tasks/{task-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Task Name",
    "trigger": {
      "type": "cron",
      "cron": "0 */30 * * * *"
    }
  }'

# Cancel/Delete a task
curl -X DELETE http://localhost:8080/api/v1/tasks/{task-id}
```

### �📚 Complete API Documentation

#### Postman Collection

- **File**: [`docs/postman_collection.json`](docs/postman_collection.json)
- **Usage**: Import into Postman for interactive API testing
- **Features**: Pre-configured requests, example payloads, response samples

## 🕒 **Task Timing Guide**

### **One-off Tasks (Execute Once)**

#### **Format**: ISO 8601 UTC datetime

```
"datetime": "YYYY-MM-DDTHH:MM:SSZ"
```

#### **Quick Examples**:

```bash
# Right now: 2025-10-02T11:46:00Z
# In 5 minutes: 2025-10-02T11:51:00Z
# In 1 hour: 2025-10-02T12:46:00Z
# Tomorrow 9 AM: 2025-10-03T09:00:00Z
# Next week: 2025-10-09T11:46:00Z
```

### **Cron Tasks (Recurring)**

#### **Format**: `second minute hour day month weekday`

```
"cron": "sec min hour day month weekday"
```

#### **Common Patterns**:

| **Description**     | **Cron Expression** | **Next Execution**             |
| ------------------- | ------------------- | ------------------------------ |
| Every minute        | `0 * * * * *`       | At :00 seconds of every minute |
| Every 30 seconds    | `*/30 * * * * *`    | Every 30 seconds               |
| Every 5 minutes     | `0 */5 * * * *`     | Every 5 minutes at :00 seconds |
| Hourly (at :00)     | `0 0 * * * *`       | Top of every hour              |
| Daily 9 AM          | `0 0 9 * * *`       | Every day at 9:00 AM           |
| Every Monday 9 AM   | `0 0 9 * * 1`       | Mondays at 9:00 AM             |
| Weekly (Sun 12 PM)  | `0 0 12 * * 0`      | Sundays at noon                |
| Monthly (1st, 9 AM) | `0 0 9 1 * *`       | 1st of every month, 9 AM       |

#### **Testing Cron (Fast Examples)**:

```bash
# Every 10 seconds
"cron": "*/10 * * * * *"

# Every minute
"cron": "0 * * * * *"

# Every 2 minutes
"cron": "0 */2 * * * *"
```

### **🧪 Testing Tips**

#### **For Cron Tasks**:

```bash
# Use frequent patterns for testing
"cron": "*/30 * * * * *"  # Every 30 seconds
"cron": "0 * * * * *"     # Every minute
```

## Development

### Project Structure

```
├── cmd/
│   └── server/                 # Application entry point
├── internal/
│   ├── api/                   # API routing and setup
│   │   └── handlers/          # HTTP handlers
│   ├── config/                # Configuration management
│   ├── db/                    # Database layer
│   │   └── migrations/        # SQL migrations
│   ├── models/                # Data models
│   ├── scheduler/             # Task scheduling logic
│   ├── services/              # Business logic
│   └── utils/                 # Utilities (logging, responses)
├── deployments/               # Docker configuration
├── docs/                      # API documentation
└── pkg/                       # Public packages
```

### Building

```bash
# Build binary
go build -o bin/task-scheduler ./cmd/server

# Build Docker image
docker build -f deployments/Dockerfile -t task-scheduler .
```

### Database Migrations

Migrations are automatically applied on application startup. Migration files are located in `internal/db/migrations/`.

## Cron Expression Format

The scheduler supports standard cron expressions with seconds:

```
┌─────────────── second (0-59)
│ ┌───────────── minute (0-59)
│ │ ┌─────────── hour (0-23)
│ │ │ ┌───────── day of month (1-31)
│ │ │ │ ┌─────── month (1-12)
│ │ │ │ │ ┌───── day of week (0-6) (Sunday to Saturday)
│ │ │ │ │ │
* * * * * *
```

Examples:

- `0 0 9 * * *` - Every day at 9:00 AM
- `0 */15 * * * *` - Every 15 minutes
- `0 0 */2 * * *` - Every 2 hours
- `0 0 9 * * 1` - Every Monday at 9:00 AM

```bash
# Check application logs
docker-compose logs app

# Verify environment variables
docker-compose config
```

### Health Check

```bash
# Check if the application is healthy
curl http://localhost:8080/health
```

Expected response:

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "task-scheduler",
    "version": "1.0.0"
  }
}
```
