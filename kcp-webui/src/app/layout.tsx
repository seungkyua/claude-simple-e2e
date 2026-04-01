import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'KCP WebUI — OpenStack 관리 콘솔',
  description: 'OpenStack 인프라를 웹에서 통합 관리하는 관리자 도구',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ko" className="dark" suppressHydrationWarning>
      <body className="min-h-screen bg-gray-950 text-gray-100" suppressHydrationWarning>
        {children}
      </body>
    </html>
  )
}
