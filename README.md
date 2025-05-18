# Feature Flags Service

This project is a feature flags management system built with Go, Gin, and MongoDB.

## Prerequisites
- Go 1.21+
- Docker (for MongoDB)

## Running the Service

1. **Start MongoDB using Docker Compose:**
   ```bash
   docker-compose up -d
   ```

2. **Set the MongoDB URI (optional, defaults to local):**
   ```bash
   export MONGODB_URI="mongodb://admin:password123@localhost:27017"
   ```

3. **Build and run the service:**
   ```bash
   go build -o main ./cmd/main.go
   ./main
   ```
   The service will be available at `http://localhost:8080`.

## API Endpoints
- `POST /api/features` - Create a new feature
- `GET /api/features/:id` - Get feature status
- `POST /api/features/:id/enable` - Enable a feature
- `POST /api/features/:id/disable` - Disable a feature
- `POST /api/features/dependencies` - Add a dependency between features

## Running Tests

1. **Ensure MongoDB is running (see above).**
2. **Run tests:**
   ```bash
   go test ./...
   ```

## Project Structure
- `cmd/` - Main application entry point
- `internal/` - Application code (handlers, services, models, repositories)

## API Documentation (Swagger)

Interactive API documentation is available via Swagger UI.

- After running the service, open your browser and go to:
  
  [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

- You can view all endpoints, schemas, and try out requests directly from the UI.

- The OpenAPI/Swagger JSON spec is available at:
  
  [http://localhost:8080/swagger/doc.json](http://localhost:8080/swagger/doc.json)

---

Feel free to contribute or open issues for improvements! 