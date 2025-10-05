# Mobile Legends Heroes API

REST API untuk mengelola data hero Mobile Legends dengan PostgreSQL database, authentication, dan Swagger documentation.

## 🚀 Features

- **PostgreSQL Database** dengan connection pooling
- **Authentication** dengan Bearer token
- **CRUD Operations** untuk heroes
- **Swagger/OpenAPI Documentation**
- **Environment Variables** untuk konfigurasi
- **CORS Support** untuk frontend
- **Auto-migration** dan initial data

## 📋 Requirements

- Go 1.21+
- PostgreSQL 12+
- Git

## 🛠️ Installation

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd mobile-legends-api
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Setup PostgreSQL Database**
   ```sql
   CREATE DATABASE heroes_db;
   CREATE USER postgres WITH PASSWORD 'password';
   GRANT ALL PRIVILEGES ON DATABASE heroes_db TO postgres;
   ```

4. **Configure environment variables**
   
   Edit file `config.env`:
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password
   DB_NAME=heroes_db
   DB_SSLMODE=disable
   SERVER_PORT=8080
   ```

## 🏃‍♂️ Running the Application

1. **Start PostgreSQL** (pastikan database sudah berjalan)

2. **Run the application**
   ```bash
   go run .
   ```

3. **Access the API**
   - **API Base URL:** `http://localhost:8080/api`
   - **Swagger UI:** `http://localhost:8080/swagger/`

## 📚 API Endpoints

### Authentication
- `POST /api/login` - Login dengan username/password
- `POST /api/logout` - Logout (Bearer token required)

### Heroes (CRUD)
- `GET /api/heroes` - Get all heroes
- `GET /api/heroes/{id}` - Get hero by ID
- `POST /api/heroes` - Create new hero (Auth required)
- `PUT /api/heroes/{id}` - Update hero (Auth required)
- `DELETE /api/heroes/{id}` - Delete hero (Auth required)

## 🔐 Authentication

### Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "12345"}'
```

### Using Bearer Token
```bash
curl -X GET http://localhost:8080/api/heroes \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🗄️ Database Schema

### Table: heroes
```sql
CREATE TABLE heroes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(100) NOT NULL,
    difficulty VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 📖 API Documentation

Swagger documentation tersedia di: `http://localhost:8080/swagger/`

## 🧪 Testing

### Test Credentials (dari config.yaml)
- **Username:** user1, **Password:** 12345
- **Username:** user2, **Password:** mahauser

### Example API Calls

1. **Get all heroes**
   ```bash
   curl http://localhost:8080/api/heroes
   ```

2. **Create new hero**
   ```bash
   curl -X POST http://localhost:8080/api/heroes \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -d '{"name": "Zilong", "role": "Fighter", "difficulty": "Mudah"}'
   ```

3. **Update hero**
   ```bash
   curl -X PUT http://localhost:8080/api/heroes/1 \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -d '{"name": "Alucard Updated", "role": "Fighter", "difficulty": "Sedang"}'
   ```

4. **Delete hero**
   ```bash
   curl -X DELETE http://localhost:8080/api/heroes/1 \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

## 🏗️ Project Structure

```
mobile-legends-api/
├── main.go           # Main application entry point
├── handlers.go       # HTTP handlers
├── models.go         # Data models
├── database.go       # Database connection and operations
├── config.yaml       # User authentication config
├── config.env        # Environment variables
├── go.mod           # Go modules
└── README.md        # Documentation
```

## 🔧 Configuration

### Environment Variables
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database username (default: postgres)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: heroes_db)
- `DB_SSLMODE` - SSL mode (default: disable)
- `SERVER_PORT` - Server port (default: 8080)

### Authentication Config
Edit `config.yaml` untuk menambah/ubah user:
```yaml
users:
  - username: user1
    password: 12345
  - username: user2
    password: mahauser
```

## 🚀 Production Deployment

1. **Set environment variables** di production server
2. **Configure PostgreSQL** dengan proper security
3. **Use HTTPS** untuk production
4. **Setup reverse proxy** (nginx/apache)
5. **Enable SSL** untuk database connection

## 📝 License

MIT License
