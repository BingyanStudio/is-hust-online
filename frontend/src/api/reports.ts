import api from './client'
import type { Report } from '@/types'

export function listReports(params: {
  site_id: string
  type?: number
  check_config_id?: string
  page?: number
  page_size?: number
}) {
  return api.get<any, Report[]>('/reports', { params })
}
