export interface Volume {
  id: string
  name: string
  status: string
  size: number
  volume_type: string
  description: string
  attachments: VolumeAttachment[]
  created_at: string
}

export interface VolumeAttachment {
  server_id: string
  device: string
}

export interface Snapshot {
  id: string
  name: string
  status: string
  volume_id: string
  size: number
  description: string
  created_at: string
}
