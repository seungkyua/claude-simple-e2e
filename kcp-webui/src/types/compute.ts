export interface Server {
  id: string
  name: string
  status: string
  flavor: Flavor
  addresses: Record<string, Addr[]>
  created: string
  updated: string
}

export interface Flavor {
  id: string
  name: string
  vcpus: number
  ram: number
  disk: number
}

export interface Addr {
  version: number
  addr: string
  type: string
}

export interface CreateServerRequest {
  name: string
  flavorId: string
  imageId: string
  networkIds?: string[]
  securityGroupIds?: string[]
  keyName?: string
  userData?: string
}
