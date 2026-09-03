# GoVisual Dashboard UI

The embedded GoVisual v2 dashboard is built with Preact, TypeScript, Tailwind CSS, Radix UI primitives, and esbuild.

## Development

Use Node.js 22 and install the committed dependency graph:

```bash
npm ci
npm run typecheck
npm run build
```

`npm run dev` watches JavaScript and TypeScript changes. Restart it after changing Tailwind classes or `src/styles.css` so the CSS is rebuilt too.

Source files live under `src/`. The production build writes `dashboard.js` and `styles.css` to `../static/`; both files are committed because the Go dashboard handler embeds them. Include regenerated assets with every UI change. CI rebuilds the assets and fails if the committed output is stale.

The main UI is organized around:

- `App.tsx` for request state, SSE updates, filters, and top-level views
- `components/RequestList.tsx` and `components/DetailPane.tsx` for request inspection
- `components/RequestComparison.tsx` and `components/RequestReplay.tsx` for request actions
- `components/Analytics.tsx`, `AgentActivity.tsx`, and `EnvironmentInfo.tsx` for secondary views
- `lib/api.ts` for dashboard API and event-stream types

The server sends request, replay-response, and top-level middleware-trace durations in milliseconds. Profiling, SQL, outbound HTTP, agent activity, and nested trace durations are Go `time.Duration` values encoded as nanoseconds; format them accordingly in the UI.
