import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { apiService } from '../services/api';
import { User } from '../types';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  updateUser: (user: Partial<User>) => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiService.login({ username, password });
          if (response.success && response.data) {
            const { token, user } = response.data;
            localStorage.setItem('token', token);
            set({ 
              user, 
              token, 
              isAuthenticated: true, 
              isLoading: false 
            });
            return true;
          } else {
            throw new Error(response.message || 'Login failed');
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Login failed';
          set({ 
            error: errorMessage, 
            isLoading: false,
            isAuthenticated: false,
            user: null,
            token: null
          });
          return false;
        }
      },

      logout: async () => {
        try {
          await apiService.logout();
          localStorage.removeItem('token');
          set({ 
            isAuthenticated: false, 
            user: null, 
            token: null 
          });
        } catch (error) {
          // Even if logout fails on server, clear local state
          localStorage.removeItem('token');
          set({ 
            isAuthenticated: false, 
            user: null, 
            token: null 
          });
        }
      },

      checkAuth: async () => {
        const token = localStorage.getItem('token');
        if (!token) {
          set({ isAuthenticated: false, user: null, token: null });
          return;
        }

        try {
          const response = await apiService.getCurrentUser();
          if (response.success && response.data) {
            set({ 
              user: response.data, 
              token, 
              isAuthenticated: true 
            });
          } else {
            localStorage.removeItem('token');
            set({ 
              isAuthenticated: false, 
              user: null, 
              token: null 
            });
          }
        } catch (error) {
          localStorage.removeItem('token');
          set({ 
            isAuthenticated: false, 
            user: null, 
            token: null 
          });
        }
      },

      updateUser: (userData: Partial<User>) => {
        const { user } = get();
        if (user) {
          set({
            user: { ...user, ...userData },
          });
        }
      },

      clearError: () => {
        set({ error: null });
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        // 恢复时检查认证状态
        if (state?.token) {
          // 异步检查认证状态
          setTimeout(() => {
            get().checkAuth();
          }, 0);
        }
      },
    }
  )
);