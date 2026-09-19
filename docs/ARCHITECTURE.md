# Architecture & Key Decisions

Live polling app. Create a poll, share a link, audience votes, everyone watching sees
results move in real time.

This document is the "key decisions" record the brief asks for. It is also the thing to
re-read before the technical interview — every choice below has a *why*, and the why is
the part that gets asked about.

---

## 1. The one idea the whole design hangs on

**Split the durable path from the hot path.**

MongoDB is the source of truth: every vote is a document, and the unique index on
`(pollId, voterKey)` is what *actually* guarantees one-person-one-vote.

Redis is the hot path: it holds the live tally, fans out updates to every connected
viewer, and cheaply rejects obvious duplicates before they ever reach Mongo.

Neither replaces the other. If Redis is wiped, the numbers are rebuilt from Mongo and
nothing is lost. If Mongo is slow, live viewers still get instant counts. That division
is the answer to "is Redis doing real work or is it decoration?"

---

## 2. Data model (MongoDB, database `pulse`)

### `users`

| Field | Type | Notes |
|---|---|---|
| `_id` | ObjectId | |
| `email` | string | lowercased + trimmed on write |
| `name` | string | 1–60 chars |
| `passHash` | string | bcrypt, cost 12 |
| `createdAt` | time | |

Index: `{ email: 1 }` **unique** — makes duplicate signup a database-level guarantee,
not a race-prone "check then insert".

### `polls`

| Field | Type | Notes |
|---|---|---|
| `_id` | ObjectId | |
| `code` | string | 8-char URL-safe share code, generated with `crypto/rand` |
| `question` | string | 3–200 chars |
| `options` | `[]{id, text}` | 2–10 options, each 1–100 chars |
| `ownerId` | ObjectId | |
| `status` | string | `open` \| `closed` |
| `multi` | bool | allow selecting more than one option |
| `closesAt` | *time | optional auto-close |
| `totals` | `map[string]int64` | durable tally snapshot, keyed by option id |
| `totalVotes` | int64 | |
| `createdAt` / `updatedAt` | time | |

Indexes: `{ code: 1 }` unique, `{ ownerId: 1, createdAt: -1 }`

The share code is random, not the ObjectId. ObjectIds are partly sequential and leak
creation time and rough volume; a random code doesn't let anyone enumerate other
people's polls.

### `votes`

| Field | Type | Notes |
|---|---|---|
| `_id` | ObjectId | |
| `pollId` | ObjectId | |
| `optionIds` | []string | |
| `voterKey` | string | `u:<userId>` if logged in, else `a:<sha256(ip \| ua \| pollId \| salt)>` |
| `createdAt` | time | |

Indexes: `{ pollId: 1, voterKey: 1 }` **unique**, `{ pollId: 1, createdAt: -1 }`

Keeping individual vote documents (rather than only a counter) costs very little and buys
an audit trail, the ability to rebuild tallies from scratch, and a path to "votes over
time" charts later.

**Anonymous voting is deliberately imperfect.** A hash of IP + user-agent stops casual
double-voting; it does not stop someone determined with a VPN, and it will group people
behind one office NAT. That's an accepted trade-off for a shareable public link, and it is
worth saying out loud rather than pretending the guard is airtight. Poll owners who want
a hard guarantee can require login.

---

## 3. What Redis actually does

Five distinct jobs, all on the critical path:

**1. Live counters.** `HINCRBY poll:{id}:counts {optionId} 1` plus `INCR poll:{id}:total`.
Reading current results is an O(1) hash read, not a Mongo aggregation over the votes
collection. This is what keeps the results endpoint fast as vote volume grows.

**2. Fan-out via pub/sub.** `PUBLISH poll:{id}:events {...}`. Every backend instance
holding WebSocket connections for that poll receives the message and pushes to its own
clients. This is the load-bearing reason Redis is here: an in-process map of connections
works perfectly on one server and breaks the instant there are two, because the voter
might be served by instance A while the viewer is connected to instance B. Pub/sub is
what makes the design horizontally scalable.

**3. Vote dedup at the edge.** `SET vote:{pollId}:{voterKey} 1 NX EX 86400` before
touching Mongo. Atomic check-and-set, so two simultaneous requests can't both pass.
Rejects the common case cheaply — but see §5, Mongo's unique index is the real guarantee.

**4. Rate limiting.** Fixed-window `INCR rl:{scope}:{ip}:{minute}` with `EXPIRE 60`,
applied to signup, login, poll creation and voting. A share link is public by definition,
so the vote endpoint needs a throttle.

**5. Live viewer presence.** A counter per poll maintained on socket connect/disconnect,
published on change, rendered as a "N watching" badge. Small feature, but it makes the
page feel alive even before votes arrive, and it exercises the same pub/sub path.

Poll metadata is also cached read-through with a short TTL, since the public poll page is
the single hottest read in the app.

### Atomic increment-and-publish

The tally update runs as one Lua script (`EVAL`): increment every selected option,
increment the total, bump a per-poll sequence number, read the whole tally back, and
publish it — in a single round trip that Redis executes atomically. Doing this as separate
commands would let two concurrent voters interleave and publish tallies that go backwards.

### Surviving a cold Redis

If `poll:{id}:counts` is missing, the tally is rebuilt from `polls.totals` in Mongo under
a short lock (`SET poll:{id}:lock NX PX 5000`), then served normally. Redis loss degrades
to one extra Mongo read. It never produces wrong numbers, and it never requires a manual
fix.

---

## 4. Key layout

```
poll:{id}:counts     hash    optionId -> count
poll:{id}:total      string  total votes
poll:{id}:seq        string  monotonic event sequence
poll:{id}:meta       string  cached poll JSON (TTL 60s)
poll:{id}:viewers    string  live socket count
poll:{id}:events     channel pub/sub topic
vote:{pollId}:{key}  string  dedup marker (TTL 24h)
rl:{scope}:{ip}:{m}  string  rate-limit window (TTL 60s)
```

---

## 5. The vote write path, exactly

1. `POST /api/polls/:code/vote` with `{ optionIds: ["o1"] }`
2. **Validate server-side**: poll exists, `status == open`, not past `closesAt`,
   `optionIds` non-empty and free of duplicates, every id belongs to *this* poll, and
   `len == 1` unless the poll allows multi-select.
3. Derive `voterKey` from the session, or from hashed IP + user-agent.
4. Redis `SET … NX` dedup marker. Already present → `409 already voted`.
5. Insert the vote document in Mongo. Duplicate-key error → `409` (**this is the
   authoritative check**; step 4 is only a fast filter, since Redis keys can expire or be
   evicted).
6. `$inc` the durable `totals` and `totalVotes` on the poll document.
7. Run the Lua script: increment counters, bump sequence, publish the new tally.
8. Respond with the tally.

If step 5 fails for any reason after step 4 succeeded, the dedup marker is deleted so the
voter can retry. Without that compensating delete, a transient Mongo blip would lock
someone out of voting for 24 hours.

Note the ordering: **durable write before the broadcast.** Publishing first would mean a
crash could show viewers a vote that was never actually recorded.

---

## 6. Realtime transport

WebSocket at `GET /ws/polls/:code`, public — the audience shouldn't need an account to
watch or vote.

**Hub structure.** A single `Hub` goroutine owns `map[pollID]*Room` and is driven purely
by channels. No mutexes anywhere, so data races are impossible by construction rather
than by discipline. Each `Room` is a set of clients plus exactly one Redis subscription:
the first client to join subscribes to `poll:{id}:events`, the last to leave unsubscribes
and the room is dropped. Subscriptions therefore track actual demand instead of growing
forever.

**Per connection**, two goroutines: a read pump (handles pongs, enforces a read deadline,
discards unexpected client frames) and a write pump. Exactly one goroutine ever writes to
a socket — `gorilla/websocket` requires this and violating it is a classic source of
corrupted frames.

**Keepalive.** Server pings every 25s and expects a pong within 30s. This is not optional
on hosted platforms: idle connections are commonly culled at 60s by the proxy in front of
the app, and without pings the socket dies silently and "realtime" quietly stops working
in production while appearing fine locally.

**Backpressure.** Each client has a bounded send buffer (32 messages). If it fills, that
client is disconnected rather than allowed to block the hub. One slow phone on bad wifi
must never stall the broadcast for everyone else.

**Correctness.** On connect the server immediately sends a full snapshot, so the UI has
real state before the first vote arrives. Every message carries a monotonic `seq`; if the
client sees a gap it refetches the snapshot over HTTP. That turns "dropped a message" from
a silent wrong-number bug into a self-healing resync.

**Client side.** Native `WebSocket`, reconnect with exponential backoff and jitter (1s
growing to a 30s cap — jitter matters so a server restart doesn't bring every client back
in the same instant), plus a resync on reconnect and on tab re-focus.

---

## 7. Authentication

Signup and login only, which is what the brief asks for. bcrypt at cost 12, then a
24-hour HS256 JWT delivered in an **httpOnly, Secure, SameSite=Lax cookie**.

httpOnly means a cross-site scripting bug can't read the token, which `localStorage`
cannot promise. `SameSite=Lax` is a straightforward CSRF defence — and it's only
straightforward because of the deployment choice below.

**Single origin.** The Go binary serves the built React app, the JSON API and the
WebSocket endpoint from one origin. No CORS configuration, no `SameSite=None`, no
preflight debugging, no separate CSRF token scheme, and the WebSocket inherits the cookie
automatically. Splitting frontend onto a CDN would be marginally faster to serve static
files and would cost a genuinely meaningful amount of the two days available in
cross-origin cookie and CORS fiddling. One origin is the right call at this size.

Other deliberate choices: login failures return one generic message so the endpoint can't
be used to discover which emails are registered; the JWT secret comes from the environment
and the server **refuses to start** if it's missing or short, so a weak default can never
reach production; passwords are never logged, returned, or included in any user payload.

Creating, closing and deleting polls requires auth and ownership. Viewing and voting are
public — that is the entire point of a share link.

---

## 8. API surface

```
POST   /api/auth/signup     {name,email,password}              -> user + cookie
POST   /api/auth/login      {email,password}                   -> user + cookie
POST   /api/auth/logout                                        -> 204
GET    /api/auth/me                                            -> user

POST   /api/polls           {question,options,multi,closesAt}  [auth]        -> poll
GET    /api/polls/mine                                         [auth]        -> polls
GET    /api/polls/:code                                        -> poll + tally + hasVoted
GET    /api/polls/:code/results                                -> tally (resync)
POST   /api/polls/:code/vote {optionIds}                       -> tally
POST   /api/polls/:code/close                                  [auth, owner] -> poll
DELETE /api/polls/:code                                        [auth, owner] -> 204

GET    /ws/polls/:code                                         -> websocket
GET    /healthz                                                -> 200
```

Validation lives in one place on the backend and every handler goes through it. Client-side
validation exists only to make the form pleasant; it is never trusted.

---

## 9. Repository layout

```
/backend
  cmd/server/main.go        entrypoint, wiring, graceful shutdown
  internal/config           env parsing, fail-fast validation
  internal/models           domain types
  internal/store            Mongo repositories + index setup
  internal/live             Redis counters, Lua scripts, pub/sub
  internal/ws               hub, room, client
  internal/auth             password hashing, JWT, middleware
  internal/httpapi          router, middleware, handlers
  internal/validate         shared input validation
/frontend
  src/                      React app (Vite)
/docs
  ARCHITECTURE.md           this file
docker-compose.yml          Mongo + Redis for local dev
Dockerfile                  multi-stage: build React, build Go, tiny final image
README.md
```

Handlers stay thin: parse, validate, delegate, respond. Business rules live in the service
layer, storage details stay inside `store` and `live`. Nothing in `internal/store` knows
what HTTP is.

---

## 10. Deployment

One Docker image containing the Go binary and the built React assets, deployed as a single
web service. Managed MongoDB and managed Redis alongside it.

Two things to verify before committing to a provider, rather than assuming:

- **The Redis free tier must allow `SUBSCRIBE` / `PUBLISH` over a real TCP connection.**
  Some serverless Redis products restrict long-lived stateful commands on their HTTP/REST
  layer. Use the RESP/TCP endpoint. Test it with `redis-cli` before building on it.
- **The host must support WebSockets, and its idle timeout must be longer than the ping
  interval.** Also check whether the free tier sleeps after inactivity — a cold start on a
  link a reviewer clicks is a bad first impression, and is worth an external uptime ping
  to keep warm.

`/healthz` checks Mongo and Redis connectivity, so the platform's health check fails loudly
if a dependency is unreachable instead of serving a broken app.

---

## 11. Deliberately not built

Worth naming, because "what did you leave out and why" is a fair interview question and
scope discipline reads better than a pile of half-finished features.

No email verification or password reset (no mail provider, and not what's being assessed).
No real-time editing of a poll's options once votes exist, since it would invalidate
existing tallies. No analytics beyond the live tally. No admin role. Anonymous vote
identity is best-effort, as discussed in §2.
