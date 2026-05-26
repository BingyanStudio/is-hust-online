export interface Site {
  id: string
  name: string
  url: string
  type: string
  description: string
  logo: string
  status: number // 0=enabled, 1=disabled
  created_at: number
}

export interface Client {
  id: string
  name: string
  location: string
  capabilities: number
  status: number // 0=online, 1=offline, 2=disabled
  token: string
  ip: string
  last_online: number
  created_at: number
  labels: string[]
}

export interface CheckConfig {
  id: string
  site_id: string
  client_id: string
  status: number // 0=enabled, 1=disabled
  check_type: number // 0=UNKNOWN, 1=HTTP, 2=PING, 4=TCP, 8=OTHER
  check_interval: string
  check_extra: unknown
}

export interface Check {
  id: string
  site_id: string
  client_id: string
  check_config_id: string
  timestamp: number
  type: number
  status: number // ErrorType enum
  result: string
  delay: number
  extra: unknown
}

export interface Report {
  site_id: string
  check_config_id: string
  timeframe: string
  type: number // 0=hourly, 1=daily, 2=monthly
  checks: number
  successes: number
  uptime: number
  avg_delay: number
}

export interface PagingMeta {
  page: number
  page_size: number
  total_pages: number
  has_more: boolean
}

export interface PaginatedResponse<T> {
  items: T[]
  paging: PagingMeta
}

export interface CreateSiteRequest {
  name: string
  url: string
  type?: string
  description?: string
  logo?: string
}

export interface UpdateSiteRequest {
  name?: string
  url?: string
  type?: string
  description?: string
  status?: number
  logo?: string
}

export interface CreateClientRequest {
  name: string
  location?: string
  capabilities?: number
  labels?: string[]
}

export interface UpdateClientRequest {
  name?: string
  location?: string
  capabilities?: number
  labels?: string[]
  status?: number
}

export interface CreateCheckConfigRequest {
  site_id: string
  client_id: string
  check_type: number
  check_interval: string
  check_extra?: unknown
}

export interface UpdateCheckConfigRequest {
  client_id?: string
  check_type?: number
  check_interval?: string
  check_extra?: unknown
  status?: number
}

export const CheckType = {
  UNKNOWN: 0,
  HTTP: 1,
  PING: 2,
  TCP: 4,
  OTHER: 8,
} as const

export const ErrorType = {
  NO_ERROR: 0,
  HTTP_UNREACHABLE: 1,
  HTTP_TIMEOUT: 2,
  HTTP_OTHER: 3,
  PING_TIMEOUT: 4,
  PING_UNREACHABLE: 5,
  TCP_TIMEOUT: 6,
  TCP_UNREACHABLE: 7,
} as const

export const ReportType = {
  HOURLY: 0,
  DAILY: 1,
  MONTHLY: 2,
} as const
