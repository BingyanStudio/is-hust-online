import api from './client'
import type {
  Client,
  PaginatedResponse,
  CreateClientRequest,
  UpdateClientRequest,
} from '@/types'

export function listClients(params?: { page?: number; page_size?: number }) {
  return api.get<any, PaginatedResponse<Client>>('/clients', { params })
}

export function getClient(id: string) {
  return api.get<any, Client>(`/clients/${id}`)
}

export function createClient(data: CreateClientRequest) {
  return api.post<any, Client>('/clients', data)
}

export function updateClient(id: string, data: UpdateClientRequest) {
  return api.put<any, Client>(`/clients/${id}`, data)
}

export function deleteClient(id: string) {
  return api.delete(`/clients/${id}`)
}
