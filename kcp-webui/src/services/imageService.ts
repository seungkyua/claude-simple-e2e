import api from './api'
import type { Image } from '@/types/image'
import type { ListResponse } from '@/types/api'

export const imageService = {
  listImages: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Image>>('/image/images', { params }),
  getImage: (id: string) =>
    api.get<Image>(`/image/images/${id}`),
  deleteImage: (id: string) =>
    api.delete(`/image/images/${id}`),
}
