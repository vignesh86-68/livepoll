# Setup — do this while the code is being written

Three things to get in place. Roughly 20 minutes, mostly waiting. Do them in this order.

---

## 1. Install Go (5 min)

Open **PowerShell** and run:

```powershell
winget install GoLang.Go
```

Then **close and reopen PowerShell** (the installer edits your PATH and the old window
won't see it), and check:

```powershell
go version
```

You want `go1.22` or newer. If `winget` isn't available, download the Windows MSI from
https://go.dev/dl/ and run it.

---

## 2. MongoDB Atlas free tier (8 min)

1. Sign up at https://www.mongodb.com/cloud/atlas/register
2. Create a **free M0 cluster**. Any provider/region is fine — pick one geographically
   near you for lower latency.
3. **Database Access** → *Add New Database User*. Username + password, role
   *Read and write to any database*. **Write the password down**, it's shown once.
4. **Network Access** → *Add IP Address* → *Allow access from anywhere* (`0.0.0.0/0`).

   > This is fine for an assignment and necessary because your home IP changes and the
   > deploy platform's IPs aren't known ahead of time. It's not what you'd do for a real
   > production system — there you'd use VPC peering or a fixed egress IP. Worth knowing
   > the difference, it's an easy interview question.

5. **Database** → *Connect* → *Drivers* → *Go*. Copy the connection string. It looks like:

   ```
   mongodb+srv://USER:PASSWORD@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
   ```

   Replace `<password>` with the real password. If your password has special characters
   they must be percent-encoded (`@` → `%40`, `#` → `%23`, and so on). Easiest fix is to
   use a password with only letters and digits.

---

## 3. Redis Cloud free tier (8 min)

1. Sign up at https://redis.io/try-free/
2. Create a **free 30MB database**.
3. From the database page copy the **public endpoint** (`host:port`) and the **password**.

Build the connection URL in this shape:

```
redis://default:YOUR_PASSWORD@YOUR_HOST:YOUR_PORT
```

If the provider requires TLS, the scheme is `rediss://` (two s's) instead.

**Important — verify pub/sub actually works before building on it.** Some hosted Redis
products restrict long-lived commands like `SUBSCRIBE` on certain tiers or over their REST
layer. There's a check script in the repo for this; instructions come with it. If pub/sub
turns out to be blocked, the fallback is another provider (Render Key Value, Railway, or
Aiven), not a redesign — the app code doesn't change.

---

## 4. Put the credentials somewhere safe

Create `backend/.env` (this file is gitignored and must **never** be committed):

```
MONGO_URI=mongodb+srv://...
REDIS_URL=redis://default:...
JWT_SECRET=
APP_ENV=development
PORT=8080
```

Generate a strong `JWT_SECRET` — in PowerShell:

```powershell
[Convert]::ToBase64String((1..48 | ForEach-Object { Get-Random -Max 256 }))
```

The server refuses to start if `JWT_SECRET` is missing or under 32 characters. That's
deliberate: a weak or defaulted signing secret is the single fastest way to make a JWT
system worthless, and failing loudly at boot is better than silently shipping it.

---

## When you're done

Tell me, and paste nothing — **don't paste the credentials into chat**. I never need to
see them; the code reads them from `.env` at runtime. If you want me to sanity-check the
format, replace the password with `xxx` first.
