import api from './client'
import type { Check, PaginatedResponse } from '@/types'

export function listChecks(params?: {
  site_id?: string
  client_id?: string
  page?: number
  page_size?: number
}) {
  return api.get<any, PaginatedResponse<Check>>('/checks', { params })
}
