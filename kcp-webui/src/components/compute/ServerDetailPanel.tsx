'use client'

import { useState, useEffect } from 'react'
import api from '@/services/api'

interface ServerDetail {
  id: string
  name: string
  status: string
  flavor: { id: string; name?: string }
  image: { id: string; name?: string }
  networks?: string
  addresses?: Record<string, Array<{ addr: string; version: number; 'OS-EXT-IPS:type'?: string }>>
  'OS-EXT-AZ:availability_zone'?: string
  'OS-EXT-STS:power_state'?: number
  'OS-EXT-STS:vm_state'?: string
  'OS-EXT-SRV-ATTR:host'?: string
  'OS-EXT-SRV-ATTR:instance_name'?: string
  'OS-DCF:diskConfig'?: string
  tenant_id?: string
  user_id?: string
  key_name?: string
  created?: string
  updated?: string
  hostId?: string
  security_groups?: Array<{ name: string }>
  locked?: boolean
  description?: string
}

interface Props {
  serverId: string
  onClose: () => void
  onAction: () => void
}

const powerStateMap: Record<number, string> = {
  0: 'NOSTATE', 1: 'Running', 3: 'Paused', 4: 'Shutdown', 6: 'Crashed', 7: 'Suspended',
}

export default function ServerDetailPanel({ serverId, onClose, onAction }: Props) {
  const [server, setServer] = useState<ServerDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState('')
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteMessage, setDeleteMessage] = useState('')

  useEffect(() => {
    api.get(`/compute/servers/${serverId}`)
      .then((res) => setServer(res.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [serverId])

  const handleAction = async (action: string) => {
    setActionLoading(action)
    try {
      await api.post(`/compute/servers/${serverId}/action`, { action })
      onAction()
      // 상태 갱신
      const res = await api.get(`/compute/servers/${serverId}`)
      setServer(res.data)
    } catch {
      // 에러는 무시 (상태가 맞지 않을 수 있음)
    } finally {
      setActionLoading('')
    }
  }

  const handleDeleteClick = () => {
    setDeleteMessage('')
    setShowDeleteConfirm(true)
  }

  const handleDeleteConfirm = async () => {
    setActionLoading('delete')
    try {
      await api.delete(`/compute/servers/${serverId}`)
      setDeleteMessage(`'${server?.name}' 서버가 삭제되었습니다.`)
      onAction()
      // 삭제 완료 메시지를 1.5초 보여준 후 패널 닫기
      setTimeout(() => {
        onClose()
      }, 1500)
    } catch {
      setDeleteMessage('서버 삭제에 실패했습니다.')
    } finally {
      setActionLoading('')
      setShowDeleteConfirm(false)
    }
  }

  if (loading) {
    return (
      <div className="fixed inset-y-0 right-0 z-50 w-[480px] border-l border-gray-700 bg-gray-900 p-6">
        <p className="text-gray-400">로딩 중...</p>
      </div>
    )
  }

  if (!server) {
    return (
      <div className="fixed inset-y-0 right-0 z-50 w-[480px] border-l border-gray-700 bg-gray-900 p-6">
        <p className="text-red-400">서버 정보를 불러올 수 없습니다</p>
        <button onClick={onClose} className="mt-4 text-sm text-gray-400 hover:text-white">닫기</button>
      </div>
    )
  }

  const formatNetworks = () => {
    if (server.networks) return server.networks
    if (!server.addresses) return '-'
    return Object.entries(server.addresses)
      .map(([net, addrs]) => `${net}=${addrs.map((a) => a.addr).join(', ')}`)
      .join('; ')
  }

  const fields: [string, string][] = [
    ['ID', server.id],
    ['Name', server.name],
    ['Status', server.status],
    ['VM State', server['OS-EXT-STS:vm_state'] || '-'],
    ['Power State', powerStateMap[server['OS-EXT-STS:power_state'] ?? 0] || 'NOSTATE'],
    ['Availability Zone', server['OS-EXT-AZ:availability_zone'] || '-'],
    ['Host', server['OS-EXT-SRV-ATTR:host'] || '-'],
    ['Instance Name', server['OS-EXT-SRV-ATTR:instance_name'] || '-'],
    ['Flavor', server.flavor?.name || server.flavor?.id || '-'],
    ['Image', server.image?.name ? `${server.image.name} (${server.image.id})` : server.image?.id || '-'],
    ['Networks', formatNetworks()],
    ['Key Name', server.key_name || '-'],
    ['Security Groups', (server.security_groups || []).map((sg) => sg.name).join(', ') || '-'],
    ['Project ID', server.tenant_id || '-'],
    ['User ID', server.user_id || '-'],
    ['Disk Config', server['OS-DCF:diskConfig'] || '-'],
    ['Locked', String(server.locked ?? false)],
    ['Created', server.created || '-'],
    ['Updated', server.updated || '-'],
  ]

  return (
    <>
      <div className="fixed inset-0 z-40 bg-black/30" onClick={onClose} />
      <div className="fixed inset-y-0 right-0 z-50 w-[480px] overflow-y-auto border-l border-gray-700 bg-gray-900">
        <div className="sticky top-0 flex items-center justify-between border-b border-gray-800 bg-gray-900 px-6 py-4">
          <h3 className="text-lg font-bold">서버 상세</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-white">X</button>
        </div>

        <div className="px-6 py-4">
          {/* 액션 버튼 */}
          <div className="mb-6 flex gap-2">
            {server.status === 'ACTIVE' && (
              <>
                <button
                  onClick={() => handleAction('stop')}
                  disabled={!!actionLoading}
                  className="rounded bg-yellow-600 px-3 py-1.5 text-xs font-medium hover:bg-yellow-700 disabled:opacity-50"
                >
                  {actionLoading === 'stop' ? '중지 중...' : '중지'}
                </button>
                <button
                  onClick={() => handleAction('reboot')}
                  disabled={!!actionLoading}
                  className="rounded bg-blue-600 px-3 py-1.5 text-xs font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  {actionLoading === 'reboot' ? '재부팅 중...' : '재부팅'}
                </button>
              </>
            )}
            {server.status === 'SHUTOFF' && (
              <button
                onClick={() => handleAction('start')}
                disabled={!!actionLoading}
                className="rounded bg-green-600 px-3 py-1.5 text-xs font-medium hover:bg-green-700 disabled:opacity-50"
              >
                {actionLoading === 'start' ? '시작 중...' : '시작'}
              </button>
            )}
            <button
              onClick={handleDeleteClick}
              disabled={!!actionLoading}
              className="rounded bg-red-600 px-3 py-1.5 text-xs font-medium hover:bg-red-700 disabled:opacity-50"
            >
              {actionLoading === 'delete' ? '삭제 중...' : '삭제'}
            </button>
          </div>

          {/* 삭제 확인 다이얼로그 */}
          {showDeleteConfirm && (
            <div className="mb-4 rounded border border-red-800 bg-red-900/20 p-4">
              <p className="mb-3 text-sm text-red-300">
                &apos;{server.name}&apos; 서버를 삭제하시겠습니까?
              </p>
              <p className="mb-3 text-xs text-gray-400">
                이 작업은 되돌릴 수 없습니다. 서버와 관련된 모든 데이터가 삭제됩니다.
              </p>
              <div className="flex gap-2">
                <button
                  onClick={handleDeleteConfirm}
                  disabled={actionLoading === 'delete'}
                  className="rounded bg-red-600 px-3 py-1.5 text-xs font-medium hover:bg-red-700 disabled:opacity-50"
                >
                  {actionLoading === 'delete' ? '삭제 중...' : '삭제 확인'}
                </button>
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="rounded bg-gray-700 px-3 py-1.5 text-xs hover:bg-gray-600"
                >
                  취소
                </button>
              </div>
            </div>
          )}

          {/* 삭제 결과 메시지 */}
          {deleteMessage && (
            <div className={`mb-4 rounded border px-4 py-2 text-sm ${
              deleteMessage.includes('실패')
                ? 'border-red-800 bg-red-900/30 text-red-400'
                : 'border-green-800 bg-green-900/30 text-green-400'
            }`}>
              {deleteMessage}
            </div>
          )}

          {/* Field / Value 테이블 */}
          <table className="w-full text-sm">
            <tbody className="divide-y divide-gray-800">
              {fields.map(([field, value]) => (
                <tr key={field}>
                  <td className="w-40 py-2 pr-4 text-gray-400">{field}</td>
                  <td className="py-2 break-all">{value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  )
}
