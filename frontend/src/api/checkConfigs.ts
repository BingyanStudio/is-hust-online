import api from './client'
import type {
  CheckConfig,
  PaginatedResponse,
  CreateCheckConfigRequest,
  UpdateCheckConfigRequest,
} from '@/types'

export function listCheckConfigs(params?: {
  site_id?: string
  client_id?: string
  page?: number
  page_size?: number
}) {
  return api.get<any, PaginatedResponse<CheckConfig>>('/check-configs', { params })
}

export function getCheckConfig(id: string) {
  return api.get<any, CheckConfig>(`/check-configs/${id}`)
}

export function createCheckConfig(data: CreateCheckConfigRequest) {
  return api.post<any, CheckConfig>('/check-configs', data)
}

export function updateCheckConfig(id: string, data: UpdateCheckConfigRequest) {
  return api.put<any, CheckConfig>(`/check-configs/${id}`, data)
}

export function deleteCheckConfig(id: string) {
  return api.delete(`/check-configs/${id}`)
}
