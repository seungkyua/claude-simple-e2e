import api from './api'
import type { AuditLog } from '@/types/audit'
import type { ListResponse } from '@/types/api'

export const auditService = {
  listLogs: (params?: {
    userId?: string
    action?: string
    resourceType?: string
    from?: string
    to?: string
    page?: number
    size?: number
  }) => api.get<ListResponse<AuditLog>>('/audit/logs', { params }),
}
