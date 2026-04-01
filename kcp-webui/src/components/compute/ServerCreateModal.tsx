'use client'

import { useState, useEffect } from 'react'
import api from '@/services/api'

interface Flavor {
  id: string
  name: string
  vcpus: number
  ram: number
  disk: number
}

interface Image {
  id: string
  name: string
  status: string
}

interface Network {
  id: string
  name: string
}

interface Props {
  onClose: () => void
  onCreated: () => void
}

export default function ServerCreateModal({ onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [flavorId, setFlavorId] = useState('')
  const [imageId, setImageId] = useState('')
  const [networkIds, setNetworkIds] = useState<string[]>([])
  const [flavors, setFlavors] = useState<Flavor[]>([])
  const [images, setImages] = useState<Image[]>([])
  const [networks, setNetworks] = useState<Network[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      api.get('/compute/flavors').then((r) => setFlavors(r.data.items || [])),
      api.get('/image/images').then((r) => setImages((r.data.items || []).filter((i: Image) => i.status === 'active'))),
      api.get('/network/networks').then((r) => setNetworks(r.data.items || [])),
    ]).catch(() => {})
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await api.post('/compute/servers', {
        name,
        flavorRef: flavorId,
        imageRef: imageId,
        networks: networkIds.length > 0
          ? networkIds.map((id) => ({ uuid: id }))
          : undefined,
      })
      onCreated()
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: { message?: string } } } }
      setError(axiosErr.response?.data?.error?.message || '서버 생성에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  const handleNetworkToggle = (id: string) => {
    setNetworkIds((prev) =>
      prev.includes(id) ? prev.filter((n) => n !== id) : [...prev, id]
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-lg rounded-lg border border-gray-700 bg-gray-900 p-6">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-bold">서버 생성</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-white">X</button>
        </div>

        {error && (
          <div className="mb-4 rounded border border-red-800 bg-red-900/30 px-4 py-2 text-sm text-red-400">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm text-gray-400">서버 이름</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              placeholder="my-server"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-400">Flavor</label>
            <select
              value={flavorId}
              onChange={(e) => setFlavorId(e.target.value)}
              required
              className="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            >
              <option value="">선택하세요</option>
              {flavors.map((f) => (
                <option key={f.id} value={f.id}>
                  {f.name} (vCPU: {f.vcpus}, RAM: {f.ram}MB, Disk: {f.disk}GB)
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-400">이미지</label>
            <select
              value={imageId}
              onChange={(e) => setImageId(e.target.value)}
              required
              className="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            >
              <option value="">선택하세요</option>
              {images.map((img) => (
                <option key={img.id} value={img.id}>{img.name}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-400">네트워크 (선택)</label>
            <div className="max-h-32 space-y-1 overflow-y-auto rounded border border-gray-700 bg-gray-800 p-2">
              {networks.length === 0 ? (
                <p className="text-xs text-gray-500">네트워크 없음</p>
              ) : (
                networks.map((n) => (
                  <label key={n.id} className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={networkIds.includes(n.id)}
                      onChange={() => handleNetworkToggle(n.id)}
                      className="rounded border-gray-600"
                    />
                    {n.name}
                  </label>
                ))
              )}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded bg-gray-700 px-4 py-2 text-sm hover:bg-gray-600"
            >
              취소
            </button>
            <button
              type="submit"
              disabled={loading}
              className="rounded bg-blue-600 px-4 py-2 text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? '생성 중...' : '생성'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
