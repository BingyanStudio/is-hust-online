import api from './client'
import type {
  Site,
  PaginatedResponse,
  CreateSiteRequest,
  UpdateSiteRequest,
} from '@/types'

export function listSites(params?: { page?: number; page_size?: number }) {
  return api.get<any, PaginatedResponse<Site>>('/sites', { params })
}

export function getSite(id: string) {
  return api.get<any, Site>(`/sites/${id}`)
}

export function createSite(data: CreateSiteRequest) {
  return api.post<any, Site>('/sites', data)
}

export function updateSite(id: string, data: UpdateSiteRequest) {
  return api.put<any, Site>(`/sites/${id}`, data)
}

export function deleteSite(id: string) {
  return api.delete(`/sites/${id}`)
}
