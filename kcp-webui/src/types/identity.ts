export interface Project {
  id: string
  name: string
  description: string
  enabled: boolean
  domain_id: string
}

export interface User {
  id: string
  name: string
  email: string
  enabled: boolean
  domain_id: string
  default_project_id: string
}
