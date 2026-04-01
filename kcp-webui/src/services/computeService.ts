import api from './api'
import type { Server, Flavor, CreateServerRequest } from '@/types/compute'
import type { ListResponse } from '@/types/api'

export const computeService = {
  listServers: (params?: { page?: number; size?: number; status?: string }) =>
    api.get<ListResponse<Server>>('/compute/servers', { params }),

  getServer: (id: string) =>
    api.get<Server>(`/compute/servers/${id}`),

  createServer: (data: CreateServerRequest) =>
    api.post<Server>('/compute/servers', data),

  deleteServer: (id: string) =>
    api.delete(`/compute/servers/${id}`),

  serverAction: (id: string, action: string) =>
    api.post(`/compute/servers/${id}/action`, { action }),

  listFlavors: () =>
    api.get<Flavor[]>('/compute/flavors'),
}
