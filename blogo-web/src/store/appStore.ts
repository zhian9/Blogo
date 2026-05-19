import { create } from 'zustand'

interface AppState {
  theme: 'dark'
  searchKeyword: string
  setSearchKeyword: (keyword: string) => void
}

export const useAppStore = create<AppState>((set) => ({
  theme: 'dark',

  searchKeyword: '',

  setSearchKeyword: (keyword) => set({ searchKeyword: keyword }),
}))
