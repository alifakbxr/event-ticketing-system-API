# Event Ticketing System API

A REST API for an Event Ticketing System built with Go and PostgreSQL, featuring JWT authentication, QR code generation, and comprehensive Swagger documentation.

## 🚀 Features

- **User Management**: Register, login, JWT authentication
- **Event Management**: Full CRUD operations (admin only)
- **Ticket System**: Purchase tickets with QR code generation
- **Admin Features**: Ticket validation, attendee management, CSV export
- **Interactive API Docs**: Complete Swagger/OpenAPI documentation

## 🛠️ Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL
- **Framework**: Gin Web Framework
- **Authentication**: JWT with bcrypt
- **Documentation**: Swagger/OpenAPI 2.0

## ⚡ Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL database (local or Neon cloud database)

### Database Setup Options

#### Option 1: Neon Cloud Database (Recommended)
1. **Create a Neon account** at [neon.tech](https://neon.tech)
2. **Create a new project** and copy the connection string
3. **Configure environment**:
```bash
cp .env.example .env
# Edit .env and set DATABASE_URL with your Neon connection string
```

#### Option 2: Local PostgreSQL
1. **Install PostgreSQL** locally
2. **Create database**:
```sql
CREATE DATABASE event_ticketing;
```
3. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your local database credentials
```

### Application Setup
1. **Clone & install**:
```bash
git clone <repository-url>
cd event-ticketing-system
go mod tidy
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. **Run the server**:
```bash
go run cmd/server/main.go
```

Server starts at `http://localhost:8080`

### 📚 API Documentation
Access interactive Swagger UI at: `http://localhost:8080/swagger/index.html`

## ⚙️ Environment Configuration

### Database Connection

The application supports two methods for database configuration:

#### Method 1: DATABASE_URL (Recommended for Neon)
```env
DATABASE_URL=postgres://username:password@host/database?sslmode=require
```

#### Method 2: Individual Variables (Fallback)
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=event_ticketing
```

### Server Configuration
```env
PORT=8080
JWT_SECRET=your-secret-key-change-this-in-production
```

## 🔑 Authentication

Use JWT tokens in Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## 👥 User Roles

- **user**: Browse events, purchase tickets, view own tickets
- **admin**: All user permissions + event management, ticket validation, attendee management

## 📱 API Usage

### Quick Examples

**Register a user**:
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
```

**Login**:
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```

**Get events**:
```bash
curl -X GET http://localhost:8080/api/events
```

**Purchase tickets**:
```bash
curl -X POST http://localhost:8080/api/events/1/purchase \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"quantity":2}'
```

## 🏗️ Project Structure

```
event-ticketing-system/
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/           # JWT authentication
│   ├── database/       # Database connection
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # Custom middleware
│   └── models/         # Database models
├── pkg/utils/          # Utility functions
└── docs/               # Swagger documentation
```

## 📄 License

MIT License