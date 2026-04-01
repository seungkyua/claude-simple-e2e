export interface Network {
  id: string
  name: string
  status: string
  subnets: string[]
  admin_state_up: boolean
  shared: boolean
}

export interface Subnet {
  id: string
  name: string
  network_id: string
  cidr: string
  ip_version: number
  gateway_ip: string
  enable_dhcp: boolean
}

export interface Router {
  id: string
  name: string
  status: string
  admin_state_up: boolean
}

export interface SecurityGroup {
  id: string
  name: string
  security_group_rules: SecurityGroupRule[]
}

export interface SecurityGroupRule {
  id: string
  direction: string
  protocol: string
  port_range_min: number | null
  port_range_max: number | null
  remote_ip_prefix: string
}
