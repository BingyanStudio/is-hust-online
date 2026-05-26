import api from './client'
import type { Report } from '@/types'

export function listReports(params: {
  site_id: string
  type?: number
  page?: number
  page_size?: number
}) {
  return api.get<any, Report[]>('/reports', { params })
}
