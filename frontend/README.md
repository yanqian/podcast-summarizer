# Frontend Skeleton

Initialized directories for Vite + React + TypeScript + Tailwind + shadcn/ui:
- `src/components`, `src/pages`, `src/hooks`, `src/services`, `src/styles`
- `tests/unit`

Current behavior:
- Accepts podcast URL, calls ingest, subscribes to SSE stream at `/api/streams/transcript/{jobId}`, renders streaming transcript chunks, then fetches aligned transcript/summaries when stream completes.
- Status banner covers idle/loading/streaming/error/success states.

Configuration:
- `VITE_API_BASE_URL` defaults to `http://localhost:8080`.
- When deploying the frontend to Vercel/Cloudflare Pages, set `VITE_API_BASE_URL` to the public HTTPS backend URL, such as a Cloudflare Tunnel or Tailscale Serve/Funnel URL. Avoid an `http://` backend from an HTTPS-hosted frontend because browsers will block mixed-content requests.

Tests:
- `npm test` runs vitest unit tests (UrlInput). E2E removed; reintroduce Playwright if needed.
