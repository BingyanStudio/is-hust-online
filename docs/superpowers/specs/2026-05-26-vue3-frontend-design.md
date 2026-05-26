# is-hust-online Vue 3 Frontend Design

## Overview

A static Vue 3 frontend for the is-hust-online distributed uptime monitoring system. Two sections: a public status page for visitors and an admin dashboard for management.

## Tech Stack

- Vue 3 + TypeScript + Vite
- vue-router (route-based auth: `/` public, `/admin` gated)
- Element Plus (UI components)
- ECharts via vue-echarts (uptime time-series charts)
- Axios (HTTP client with response interceptor)

## Routes

| Path | View | Auth |
|---|---|---|
| `/` | Public status page — site list | No |
| `/:id` | Site detail — uptime chart + check history | No |
| `/admin` | Redirects to `/admin/sites` | Yes (browser Basic Auth) |
| `/admin/sites` | Sites CRUD | Yes |
| `/admin/clients` | Clients CRUD | Yes |
| `/admin/check-configs` | CheckConfigs CRUD | Yes |

## Public Pages

### Status Page (`/`)

List of all enabled sites. Each row displays:
- Status indicator (green = online/enabled, red = offline/disabled)
- Site name
- Site URL
- Uptime percentage

Click a row → navigates to site detail.

### Site Detail (`/:id`)

- **Uptime chart**: ECharts line chart showing uptime over time. Granularity selector (hourly / daily / monthly). Timeframe date range picker.
- **Check history table**: Recent check results — timestamp, status (success/error type), response delay (ms). Paginated.

## Admin Pages

### Layout

Left sidebar with navigation:
- Sites
- Clients
- Check Configs

Active item highlighted. Main content area to the right.

### Sites Management

Data table with columns: Name, URL, Status (enabled/disabled badge), Created At. Actions: Edit, Delete. "Add Site" button opens create dialog.

**Create/Edit dialog fields**: name (required), url (required), type, description, logo, status.

### Clients Management

Data table with columns: Name, Location, Status (online/offline/disabled badge), Labels, Last Online. Token and IP shown only when authenticated (backend handles redaction). Actions: Edit, Delete. "Add Client" button opens create dialog.

**Create/Edit dialog fields**: name (required), location, capabilities, labels, status.

### Check Configs Management

Data table with columns: Site (name), Client (name), Check Type, Interval (cron), Status. Actions: Edit, Delete. "Add Check Config" button opens create dialog.

**Create/Edit dialog fields**: site_id (required, dropdown), client_id (required, dropdown), check_type (required, select), check_interval (required, cron string), check_extra, status.

## API Layer

### Axios Client

```typescript
const api = axios.create({ baseURL: '/api' })

api.interceptors.response.use((res) => {
  if (res.data.code !== 0) return Promise.reject(res.data.message)
  return res.data.data
})
```

All API functions return unwrapped data. Errors reject with the message string.

### API Modules

Each entity has its own module (`api/sites.ts`, `api/clients.ts`, etc.) exporting functions like:
- `listSites(params)` → `{ items: Site[], paging: PagingMeta }`
- `getSite(id)` → `Site`
- `createSite(data)` → `Site`
- `updateSite(id, data)` → `Site`
- `deleteSite(id)` → void

### Auth

Browser handles Basic Auth natively. When a 401 response is received, the browser shows the native username/password prompt. No login page needed.

### Dev Proxy

Vite config proxies `/api` to the Go backend (default port from config).

## Component Structure

```
src/
├── App.vue
├── main.ts
├── router/index.ts
├── api/
│   ├── client.ts
│   ├── sites.ts
│   ├── clients.ts
│   ├── checkConfigs.ts
│   ├── checks.ts
│   └── reports.ts
├── types/index.ts
├── views/
│   ├── public/
│   │   ├── StatusPage.vue
│   │   └── SiteDetail.vue
│   └── admin/
│       ├── AdminLayout.vue
│       ├── Sites.vue
│       ├── Clients.vue
│       └── CheckConfigs.vue
└── components/
    ├── UptimeChart.vue
    ├── CheckHistoryTable.vue
    └── SiteStatusBadge.vue
```

## State Management

No Pinia. Each view manages its own state with `ref()`. Data fetched in `onMounted()`. Child components receive data via props.

## Error Handling

- API errors (non-zero code) reject the axios interceptor promise
- Views catch errors and show Element Plus `ElMessage.error()` notifications
- Loading states shown via `v-loading` directive on tables

## No Testing

No test suite is planned for this initial version.
