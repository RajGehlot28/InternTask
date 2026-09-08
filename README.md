# Ticket System Backend API in Golang

A backend service built in **Golang** running on port **8080**, featuring JWT authentication, password hashing, user-based authorization, ticket status workflow management, and a PostgreSQL database layer, paired with an HTML/CSS/JS frontend interface.

---

## Features

- **User Authentication**: User registration and login with bcrypt password hashing.
- **JWT Security**: Protected REST endpoints using `Authorization: Bearer <token>`.
- **Ownership Authorization**: Users can view and update only their own tickets.
- **Status Flow**: Enforces state transitions (`open` -> `in_progress` -> `closed`). Prevents reopening closed tickets.
- **PostgreSQL Database**: Direct PostgreSQL integration (`DATABASE_URL`).
- **Health Check API**: Public `/health` endpoint returning `{"status": "ok"}`.
- **Frontend UI**: Single-page dashboard interface.
- **Dockerized**: Multi-stage `Dockerfile` ready for deployment.

---

## Project Architecture

```text
               User / Browser
                     │
                     ▼
           Frontend (HTML/CSS/JS)
                     │
                     ▼
      Golang REST API (Port 8080)
 ┌───────────────────┴───────────────────┐
 │                                       │
 ▼                                       ▼
Public Endpoints                Protected Endpoints
 ├── GET  /health                ├── POST  /tickets
 ├── POST /auth/register         ├── GET   /tickets
 └── POST /auth/login            ├── GET   /tickets/{id}
                                 └── PATCH /tickets/{id}/status
                                         │
                                         ▼
                               JWT Authorization &
                           Ownership Security Check
                                         │
                                         ▼
                                PostgreSQL Database
```

---

## Repository Structure

```text
InternTask/
├── backend/
│   ├── auth.go          # JWT token generation & auth middleware
│   ├── db.go            # PostgreSQL connection & migrations
│   ├── handlers.go      # REST API route handlers
│   ├── go.mod           # Go module file
│   ├── go.sum           # Go module checksums
│   └── main.go          # Server entrypoint
├── frontend/
│   ├── css/
│   │   └── style.css    # Frontend CSS styles
│   ├── js/
│   │   └── app.js       # Frontend JS logic
│   └── index.html       # Single-page app HTML
├── Dockerfile           # Multi-stage Docker build file
├── .env.example         # Environment variables template
├── .gitignore          # Git ignore rules
└── README.md            # Project documentation
```

---

## API Endpoints

| Method | Endpoint | Access | Description |
|---|---|---|---|
| `GET` | `/health` | Public | Returns health status `{"status": "ok"}` |
| `POST` | `/auth/register` | Public | Register user & return JWT token |
| `POST` | `/auth/login` | Public | Login user & return JWT token |
| `POST` | `/tickets` | Protected | Create a new ticket owned by logged-in user |
| `GET` | `/tickets` | Protected | List all tickets belonging to logged-in user |
| `GET` | `/tickets/{id}` | Protected | Get ticket by ID (returns `403` if not owned) |
| `PATCH` | `/tickets/{id}/status` | Protected | Update ticket status (`open` -> `in_progress` -> `closed`) |

### Status Flow Rules:
1. Supported statuses: `open`, `in_progress`, `closed`.
2. Valid sequence: `open` -> `in_progress` -> `closed` (or `open` -> `closed`).
3. Closed tickets cannot move back to `open` or `in_progress`.

---

## Environment Variables (.env)

```env
PORT=8080
JWT_SECRET=supersecretkey123
DATABASE_URL=postgres://username:password@hostname:5432/dbname?sslmode=require
```

---

## Local Run Instructions

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Run the application:
   ```bash
   go run .
   ```
   The service will start on `http://localhost:8080`.

3. Test the Health Check endpoint:
   ```bash
   curl http://localhost:8080/health
   ```

---

## Docker Run Instructions

1. Build the Docker image from the root directory:
   ```bash
   docker build -t ticket-system .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 -e DATABASE_URL="postgres://..." ticket-system
   ```
