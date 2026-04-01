import { create } from 'zustand'

interface AuthState {
  token: string | null
  username: string | null
  isAuthenticated: boolean
  login: (token: string, username: string) => void
  logout: () => void
}

/** 인증 상태를 관리하는 Zustand 스토어 */
export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  username: null,
  isAuthenticated: false,

  login: (token, username) =>
    set({ token, username, isAuthenticated: true }),

  logout: () =>
    set({ token: null, username: null, isAuthenticated: false }),
}))
