# Multi-Window Media Sequencer with Sync Playback

> A self-healing, multi-window media playback engine and real-time broadcast synchronizer built with **Go 1.24** and **React + Vite**.

---

## 1. Overview (Plain-Language Explanation)

Think of each "window" in this application like its own dedicated **TV Channel**. 

Each channel follows its own scheduled program of media clips (images, videos, or configured blank slots) that repeats continuously. The entire sequence is designed to fit inside a **5-hour loop** — after 5 hours, the channel automatically restarts from the top of its schedule.

### What is "Sync"?
"Sync" acts like an **Emergency Broadcast Override**:
1. When a user triggers "Sync All" with a specific media clip (e.g. `m2`), **every window on every connected device instantly switches to play that clip simultaneously**.
2. A countdown timer runs for the specified sync duration (e.g., 10 seconds).
3. The moment the sync period ends, every channel **automatically switches back to what it would have been playing anyway** on its normal schedule, as if nothing happened.

---

## 2. Core Architectural Insight: Deterministic "Virtual Clock" Playback

### The Problem with the Naive Approach
A naive implementation uses a background loop or mutable pointer (`window_3 is at item_2, 12 seconds in`). If the server restarts, a tab reconnects mid-cycle, or a broadcast sync interrupts playback, you must manually repair and stitch pointers. This leads to common bugs: videos restarting from 0s, playlists drifting out of sync across windows, or corrupt sequence states.

### The Virtual Clock Solution
Instead of storing or mutating "what is currently playing", playback position at any instant is a **pure, stateless mathematical calculation derived from the clock**:

$$\text{elapsedInCycle} = (\text{now} - \text{cycleEpoch}) \bmod \text{cycleLength}$$
$$\text{positionInLoop} = \text{elapsedInCycle} \bmod \text{playlistTotalDuration}$$

We then walk the playlist's cumulative item durations to find which media item contains `positionInLoop` and compute the sub-second `offset_seconds`.

### Why This Design Solves Every Requirement
- **Self-Healing & Stateless**: Refreshing the browser or restarting the Go server recalculates the exact same playback position instantly.
- **Seamless Sync Resume**: Because we never pause or mutate underlying playlist state during broadcast sync, when sync ends, the formula simply resumes giving the current schedule point without needing stitching logic.
- **Zero Drift**: Client clocks align with the server clock via an initial health check time-offset measurement (`useServerTimeOffset`).

---

## 3. Tech Stack

| Component | Choice | Reason |
| :--- | :--- | :--- |
| **Backend Language** | Go 1.24 | High performance, lightweight concurrency, strong standard library |
| **HTTP Router** | Chi (`go-chi/chi/v5`) | Idiomatic, lightweight Go HTTP router |
| **Database** | SQLite (`modernc.org/sqlite`) | CGO-free pure Go SQLite driver (zero C toolchain required for Docker) |
| **Realtime Streaming** | Server-Sent Events (SSE) | Lightweight 1-way push via standard `http.Flusher` |
| **Frontend UI** | React 18 + Vite | Fast rendering, modular component hierarchy |
| **Styling** | Modern CSS System | Dark mode, glassmorphic cards, vibrant glows, micro-animations |

---

## 4. API Documentation

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Server health check & reference timestamp for clock skew correction |
| `GET` | `/windows` | List all configured windows with ordered playlists |
| `GET` | `/windows/{id}` | Retrieve details for a single window |
| `GET` | `/windows/{id}/now` | Server-computed "what's playing right now" for debug/convenience |
| `POST` | `/windows` | Create a new window |
| `POST` | `/windows/{id}/media` | Append a media item to a window's playlist (by `media_id` or inline creation) |
| `GET` | `/media` | List all items in the shared media library |
| `POST` | `/media` | Create a new media item (type: `image`, `video`, or `blank`) |
| `POST` | `/sync` | Trigger global broadcast sync override across all connected windows |
| `GET` | `/sync/status` | Current global sync override state |
| `GET` | `/events` | Real-time Server-Sent Events stream (`sync` and `playlist_updated` push events) |

### API Payloads & Examples

#### `GET /windows/w1/now`
```json
{
  "window_id": "w1",
  "media": {
    "id": "m2",
    "type": "video",
    "url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4",
    "duration_seconds": 30,
    "created_at": "2026-09-09T17:39:37Z"
  },
  "offset_seconds": 14.85,
  "is_sync": false
}
```

#### `POST /sync`
```json
// Request
{
  "media_id": "m2",
  "duration_seconds": 10
}

// Response (200 OK)
{
  "active": true,
  "media_id": "m2",
  "media": { ... },
  "started_at": "2026-09-09T17:42:50Z",
  "duration_seconds": 10,
  "ends_at": "2026-09-09T17:43:00Z"
}
```

#### `POST /windows/w1/media`
```json
// Request (Reference Existing Library Item)
{ "media_id": "m1" }

// Request (Inline Custom Item Creation)
{
  "type": "image",
  "url": "https://images.unsplash.com/photo-1550684848-fac1c5b4e853?w=1200",
  "duration_seconds": 25
}
```

---

## 5. Local Setup & Execution Instructions

### Prerequisites
- **Go**: 1.24+
- **Node.js**: v18+ & `npm`

### 1. Run Backend Server
```bash
cd backend
go run cmd/server/main.go
```
- Server starts on `http://localhost:8085`.
- Database (`sequencer.db`) is automatically created and seeded with sample windows (`w1`, `w2`) and media library items (`m1`, `m2`, `m3` blank slot, `m4`).

### 2. Run Frontend Web App
```bash
cd frontend
npm install
npm run dev
```
- Web application starts on `http://localhost:5173`.
- Open `http://localhost:5173` in multiple browser windows or tabs to observe synchronized playback in real time!

---

## 6. Deployment & Docker Containerization

### Option A: Docker Compose (Full Stack - Recommended)
Run both the Go backend and React frontend via Nginx with a single command:
```bash
docker-compose up --build -d
```
- **Frontend App**: `http://localhost:3000`
- **Backend API**: `http://localhost:8085`

### Option B: Standalone Backend Container
```bash
cd backend
docker build -t media-sequencer-backend .
docker run -p 8085:8080 -d --name media-sequencer media-sequencer-backend
```
Verify container health:
```bash
curl http://localhost:8085/health
```

### Continuous Integration (CI/CD)
This repository includes a GitHub Actions pipeline (`.github/workflows/deploy.yml`) that automatically runs Go tests, builds the Vite production bundle, and validates Docker builds on every commit to `main`.


---

## 7. Assumptions & Technical Tradeoffs

1. **Video Duration**: Since media items can be external URLs, `duration_seconds` is specified at creation time rather than auto-extracted server-side.
2. **Sync Precision**: Realtime SSE pushes deliver sync events to clients within normal network latency (milliseconds). Sub-millisecond physical clock synchronization across independent hardware monitors is bounded by device network clocks.
3. **CGO-Free SQLite**: Chosen `modernc.org/sqlite` pure-Go driver over C-dependent drivers so Docker builds require zero C toolchains and deploy cleanly to any environment.
4. **Blank Slot Handling**: Blank slots are modeled explicitly as media items with `type: "blank"` and `url: null`, ensuring blank states only render when deliberately scheduled in a window's playlist loop.
