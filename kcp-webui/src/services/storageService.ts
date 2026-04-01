import api from './api'
import type { Volume, Snapshot } from '@/types/storage'
import type { ListResponse } from '@/types/api'

export const storageService = {
  listVolumes: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Volume>>('/storage/volumes', { params }),
  createVolume: (data: { name: string; size: number; volumeType?: string }) =>
    api.post<Volume>('/storage/volumes', data),
  deleteVolume: (id: string) =>
    api.delete(`/storage/volumes/${id}`),
  attachVolume: (id: string, serverId: string) =>
    api.post(`/storage/volumes/${id}/attach`, { serverId }),
  detachVolume: (id: string) =>
    api.post(`/storage/volumes/${id}/detach`),

  listSnapshots: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Snapshot>>('/storage/snapshots', { params }),
  createSnapshot: (data: { name: string; volumeId: string }) =>
    api.post<Snapshot>('/storage/snapshots', data),
  deleteSnapshot: (id: string) =>
    api.delete(`/storage/snapshots/${id}`),
}
