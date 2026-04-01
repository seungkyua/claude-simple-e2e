import api from './api'
import type { Network, Subnet, Router, SecurityGroup } from '@/types/network'
import type { ListResponse } from '@/types/api'

export const networkService = {
  listNetworks: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Network>>('/network/networks', { params }),
  createNetwork: (data: { name: string; adminStateUp?: boolean; shared?: boolean }) =>
    api.post<Network>('/network/networks', data),
  deleteNetwork: (id: string) =>
    api.delete(`/network/networks/${id}`),

  listSubnets: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Subnet>>('/network/subnets', { params }),
  createSubnet: (data: { name: string; networkId: string; cidr: string; ipVersion?: number }) =>
    api.post<Subnet>('/network/subnets', data),
  deleteSubnet: (id: string) =>
    api.delete(`/network/subnets/${id}`),

  listRouters: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<Router>>('/network/routers', { params }),
  createRouter: (data: { name: string }) =>
    api.post<Router>('/network/routers', data),
  deleteRouter: (id: string) =>
    api.delete(`/network/routers/${id}`),

  listSecurityGroups: (params?: { page?: number; size?: number }) =>
    api.get<ListResponse<SecurityGroup>>('/network/security-groups', { params }),
  createSecurityGroup: (data: { name: string; description?: string }) =>
    api.post<SecurityGroup>('/network/security-groups', data),
  deleteSecurityGroup: (id: string) =>
    api.delete(`/network/security-groups/${id}`),
}
