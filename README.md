# LivePoll — Real-time Polling App

A live polling tool where users create polls, share a link, and audiences vote — with results updating in real-time, no refresh needed.

**Live:** [https://livepoll-l0lk.onrender.com](https://livepoll-l0lk.onrender.com)

## Tech Stack

| Layer | Technology | Role |
|---|---|---|
| Frontend | React (Vite) | SPA with live-updating UI |
| Backend | Go (Gin) | REST API + WebSocket server |
| Database | MongoDB | Source of truth — users, polls, votes |
| Realtime | Redis | Live counters, pub/sub fan-out, vote dedup, rate limiting |

## How It Works

1. **Sign up / Log in** → create a poll with 2–10 options
2. **Share the link** → anyone can vote without an account
3. **Results update live** → WebSocket pushes every vote to every viewer instantly
4. **One vote per person** → Redis dedup (fast path) + MongoDB unique index (authoritative)

### Why Both MongoDB and Redis?

MongoDB is the durable source of truth. Redis is the hot path: it holds live tallies, fans out updates via pub/sub to every connected WebSocket, and rejects duplicate votes at the edge before they reach Mongo. If Redis is wiped, tallies rebuild from Mongo automatically. Neither replaces the other.

## Key Decisions

- **Single origin** — Go serves the built React app, the API, and WebSocket from one origin. No CORS, no cross-origin cookie issues.
- **JWT in httpOnly cookie** — XSS can't read it. `SameSite=Lax` handles CSRF.
- **Lua script for atomic tally** — increment counters + publish update in one Redis round-trip, no interleaving.
- **Channel-driven WebSocket hub** — no mutexes, data races impossible by construction.
- **Durable write before broadcast** — vote hits Mongo before Redis publishes, so a crash can't show phantom votes.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full design rationale.

## Run Locally

### Prerequisites

- Go 1.22+
- Node.js 18+
- MongoDB Atlas free tier (or local via Docker)
- Redis Cloud free tier (or local via Docker)

### Setup

1. Clone the repo:
   ```bash
   git clone https://github.com/vignesh/livepoll.git
   cd livepoll
   ```

2. Create `backend/.env`:
   ```
   MONGO_URI=mongodb+srv://...
   REDIS_URL=redis://default:...
   JWT_SECRET=<at-least-32-characters>
   APP_ENV=development
   PORT=8080
   ```

3. Install frontend dependencies:
   ```bash
   cd frontend && npm install && cd ..
   ```

4. Start the Go backend:
   ```bash
   cd backend && go run ./cmd/server
   ```

5. In another terminal, start the Vite dev server:
   ```bash
   cd frontend && npm run dev
   ```

6. Open http://localhost:5173

### Production Build

```bash
cd frontend && npm run build
# The Go server serves the built files from ./web by default
cp -r frontend/dist backend/web
cd backend && go run ./cmd/server
# Now open http://localhost:8080
```

## Deploy

The app deploys as a single Docker image:

```bash
docker build -t livepoll .
docker run -p 8080:8080 \
  -e MONGO_URI="..." \
  -e REDIS_URL="..." \
  -e JWT_SECRET="..." \
  -e APP_ENV=production \
  -e VOTER_SALT="..." \
  livepoll
```

Tested on **Render** (free Web Service with Docker). Connect the GitHub repo, set environment variables, and deploy.

## Project Structure

```
/backend
  cmd/server/main.go        → entrypoint, wiring, graceful shutdown
  internal/config            → env parsing, fail-fast validation
  internal/models            → domain types (User, Poll, Vote, Tally)
  internal/store             → MongoDB repositories + index setup
  internal/live              → Redis counters, Lua scripts, pub/sub, rate limiting
  internal/ws                → WebSocket hub, rooms, clients
  internal/auth              → bcrypt hashing, JWT issue/verify
  internal/httpapi           → Gin router, middleware, handlers
  internal/validate          → input validation
  internal/idgen             → cryptographic share code generation
/frontend
  src/                       → React app (Vite)
/docs
  ARCHITECTURE.md            → full design rationale + key decisions
Dockerfile                   → multi-stage build
docker-compose.yml           → local Mongo + Redis (convenience)
```

## Features

- ✅ Real-time vote updates via WebSocket (no refresh)
- ✅ Authentication (signup/login with bcrypt + JWT)
- ✅ Backend validation on all inputs
- ✅ One-vote-per-person (Redis dedup + MongoDB unique index)
- ✅ Multi-select polls
- ✅ Auto-close polls at a deadline
- ✅ Live "N watching" viewer count
- ✅ Share link (copy to clipboard)
- ✅ Rate limiting on public endpoints
- ✅ Responsive dark-mode UI with glassmorphism
- ✅ Graceful shutdown
- ✅ Health check endpoint (`/healthz`)
