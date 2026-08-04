<p align="center">
  <img src="docs/assets/banner.svg" alt="google-health-mcp — sleep · activity · recovery" width="100%">
</p>

# google-health-mcp

Static Go [MCP](https://modelcontextprotocol.io) for **Google Health** data
from modern **Fitbit / Pixel Watch** devices via the
[Google Health API](https://developers.google.com/health/about)
(`health.googleapis.com`).

Built for [ai-gantry](https://github.com/shotah/ai-gantry) / LOCAL_AGENT: lean
**core** tool surface (Garmin-shaped nouns), stdio only, `CGO_ENABLED=0`.
**Read + summarize — never writes fake workouts.**

Naming follows
[ai-gantry MCP tool naming](https://github.com/shotah/ai-gantry/blob/main/docs/mcp-naming.md):
host server id `ghealth` → tools like `ghealth__sleep_get` (tool names do
**not** repeat the server id).

> **Local-only scaffold.** No `git init`, no GitHub repo, no push from this
> agent. Create the repo under your **personal** account when ready — see
> [TODO.md](TODO.md).

## New school only

| Path | This package? |
| --- | --- |
| **Google Health API** + Google OAuth 2.0 | ✅ Yes |
| Legacy Fitbit Web API (FOT auth) | ❌ Turned down Sept 2026 — skip |
| Google Fit REST | ❌ Closed to new apps — skip |
| Health Connect (on-device Android) | ❌ Wrong shape for stdio MCP |

Friends on the new Google/Fitbit line sync through the Fitbit app → data is
available on **Google Health API**. Chris keeps [`go-garmin`](https://github.com/shotah/go-garmin).

Background: [docs/why-google-health.md](docs/why-google-health.md).

## Tools (planned core tier)

| Tool | Google Health data type(s) |
| --- | --- |
| `sleep_get` | `sleep` (reconcile) |
| `activities_list` | `exercise` (list / reconcile) |
| `activities_get` | `exercise` get by name |
| `heart_rate_get` | `daily-resting-heart-rate` (+ optional `heart-rate`) |
| `hrv_get` | `daily-heart-rate-variability` |
| `weight_get` | `weight` |
| `profile_get` | identity / profile |
| `account_get` | auth / token status |

Host: `ghealth__sleep_get`, `ghealth__activities_list`, …

### Agent contract

```text
sleep_get / heart_rate_get / hrv_get  → recovery
activities_list → activities_get     → “what did I do?”
weight_get                           → scale
```

Same recipes as `garmin__…` — swap prefix for Fitbit friends.

## Setup

1. Personal GCP project → enable **Google Health API** → OAuth Web client  
   ([setup](https://developers.google.com/health/setup)).
2. Add **readonly** scopes you need on the Data Access page (Restricted —
   testing limited to ~100 users until verification).
3. Export:

```bash
export GOOGLE_HEALTH_CLIENT_ID=...
export GOOGLE_HEALTH_CLIENT_SECRET=...
# after auth:
# export GOOGLE_HEALTH_TOKEN_PATH=...
```

4. Auth (planned): `google-health-mcp auth` → Google OAuth → tokens on disk.

### Core scopes (readonly)

```text
https://www.googleapis.com/auth/googlehealth.sleep.readonly
https://www.googleapis.com/auth/googlehealth.activity_and_fitness.readonly
https://www.googleapis.com/auth/googlehealth.health_metrics_and_measurements.readonly
https://www.googleapis.com/auth/googlehealth.profile.readonly
```

### Boot safety

MCP **starts even if tokens are missing**. Data tools return a clear auth
error. Optional fitness MCP must not take Tim down.

## Development

```bash
make help
make tools
make check
make cli
make self-test
```

Status: **docs + compileable stub** targeting `health.googleapis.com/v4`.
Live client + OAuth in [TODO.md](TODO.md).

## ai-gantry wiring (when published)

```toml
[[server]]
name = "ghealth"
command = "google-health-mcp"
auth_args = ["auth"]
download_tag = "latest"
download_url = "https://github.com/shotah/google-health-mcp/releases/download/{tag}/google-health-mcp_{version}_{os}_{arch}.tar.gz"
```

## License

MIT
