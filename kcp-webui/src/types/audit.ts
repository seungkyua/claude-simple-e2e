export interface AuditLog {
  id: string
  user: { id: string; username: string }
  action: string
  resourceType: string
  resourceId: string
  source: string
  statusCode: number
  requestSummary: string
  ipAddress: string
  createdAt: string
}
