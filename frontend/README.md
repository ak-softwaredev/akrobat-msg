# Akrobat frontend

React and TypeScript client for the Akrobat messaging API.

## Run locally

Start the Go API from the repository root with `make run`, then start Vite from this directory:

```sh
npm install
npm run dev
```

Open `http://localhost:5173`. Vite proxies `/api` requests and WebSocket upgrades to `https://localhost:8080`, accepting the repository's local development certificate. The API origin can be overridden with `VITE_API_URL` when the deployment provides the required cross-origin access or a same-origin reverse proxy.

## Build

```sh
npm run build
npm run lint
```
