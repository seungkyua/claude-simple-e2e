interface StatCardProps {
  title: string
  value: number | string
  subtitle?: string
  color?: 'blue' | 'green' | 'yellow' | 'red'
}

const colorMap = {
  blue: 'border-blue-500 text-blue-400',
  green: 'border-green-500 text-green-400',
  yellow: 'border-yellow-500 text-yellow-400',
  red: 'border-red-500 text-red-400',
}

export default function StatCard({ title, value, subtitle, color = 'blue' }: StatCardProps) {
  return (
    <div className={`rounded-lg border-l-4 bg-gray-900 p-4 ${colorMap[color]}`}>
      <p className="text-sm text-gray-400">{title}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
      {subtitle && <p className="mt-1 text-xs text-gray-500">{subtitle}</p>}
    </div>
  )
}
