import api from './api'
import type { Project, User } from '@/types/identity'
import type { ListResponse } from '@/types/api'

export const identityService = {
  listProjects: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Project>>('/identity/projects', { params }),
  createProject: (data: { name: string; description?: string }) =>
    api.post<Project>('/identity/projects', data),
  deleteProject: (id: string) =>
    api.delete(`/identity/projects/${id}`),

  listUsers: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<User>>('/identity/users', { params }),
  createUser: (data: { name: string; email?: string; password: string }) =>
    api.post<User>('/identity/users', data),
  deleteUser: (id: string) =>
    api.delete(`/identity/users/${id}`),

  assignRole: (data: { userId: string; projectId: string; roleId: string }) =>
    api.post('/identity/roles/assign', data),
  revokeRole: (data: { userId: string; projectId: string; roleId: string }) =>
    api.delete('/identity/roles/revoke', { data }),
}
