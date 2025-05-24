import axios, { AxiosInstance, AxiosResponse } from 'axios'

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = localStorage.getItem('refreshToken')
        if (refreshToken) {
          const response = await axios.post('/api/auth/refresh', {
            refreshToken,
          })
          
          const { accessToken } = response.data
          localStorage.setItem('token', accessToken)
          
          // Retry original request with new token
          originalRequest.headers.Authorization = `Bearer ${accessToken}`
          return api(originalRequest)
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        localStorage.removeItem('token')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
      }
    }

    return Promise.reject(error)
  }
)

// API response types
interface LoginResponse {
  accessToken: string
  refreshToken: string
  expiresAt: string
  user: {
    id: string
    username: string
    email: string
    firstName: string
    lastName: string
    isActive: boolean
    roles: string[]
  }
}

interface User {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  isActive: boolean
  roles: string[]
}

interface Domain {
  id: string
  name: string
  documentRoot: string
  isActive: boolean
  hasSSL: boolean
  phpVersion: string
  diskUsage: number
  bandwidthUsage: number
  createdAt: string
}

interface EmailAccount {
  id: string
  username: string
  quotaMB: number
  usedMB: number
  isActive: boolean
  createdAt: string
}

interface Database {
  id: string
  name: string
  type: string
  sizeMB: number
  createdAt: string
}

// Auth API
export const authAPI = {
  login: async (credentials: {
    username: string
    password: string
    twoFactorCode?: string
  }): Promise<LoginResponse> => {
    const response: AxiosResponse<LoginResponse> = await api.post('/auth/login', credentials)
    return response.data
  },

  register: async (userData: {
    username: string
    email: string
    password: string
    firstName: string
    lastName: string
  }): Promise<User> => {
    const response: AxiosResponse<User> = await api.post('/auth/register', userData)
    return response.data
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout')
  },

  refreshToken: async (refreshToken: string): Promise<LoginResponse> => {
    const response: AxiosResponse<LoginResponse> = await api.post('/auth/refresh', {
      refreshToken,
    })
    return response.data
  },
}

// User API
export const userAPI = {
  getProfile: async (): Promise<User> => {
    const response: AxiosResponse<User> = await api.get('/users/profile')
    return response.data
  },

  updateProfile: async (updates: Partial<User>): Promise<User> => {
    const response: AxiosResponse<User> = await api.patch('/users/profile', updates)
    return response.data
  },

  changePassword: async (data: {
    currentPassword: string
    newPassword: string
  }): Promise<void> => {
    await api.post('/users/change-password', data)
  },
}

// Domain API
export const domainAPI = {
  getDomains: async (page = 1, limit = 10): Promise<{
    domains: Domain[]
    total: number
    page: number
    limit: number
  }> => {
    const response = await api.get('/domains', {
      params: { page, limit },
    })
    return response.data
  },

  createDomain: async (name: string): Promise<Domain> => {
    const response: AxiosResponse<Domain> = await api.post('/domains', { name })
    return response.data
  },

  getDomain: async (id: string): Promise<Domain> => {
    const response: AxiosResponse<Domain> = await api.get(`/domains/${id}`)
    return response.data
  },

  updateDomain: async (id: string, updates: Partial<Domain>): Promise<Domain> => {
    const response: AxiosResponse<Domain> = await api.patch(`/domains/${id}`, updates)
    return response.data
  },

  deleteDomain: async (id: string): Promise<void> => {
    await api.delete(`/domains/${id}`)
  },
}

// Email API
export const emailAPI = {
  getEmailAccounts: async (domainId: string): Promise<EmailAccount[]> => {
    const response: AxiosResponse<EmailAccount[]> = await api.get(`/domains/${domainId}/email`)
    return response.data
  },

  createEmailAccount: async (domainId: string, data: {
    username: string
    password: string
    quotaMB: number
  }): Promise<EmailAccount> => {
    const response: AxiosResponse<EmailAccount> = await api.post(`/domains/${domainId}/email`, data)
    return response.data
  },

  deleteEmailAccount: async (accountId: string): Promise<void> => {
    await api.delete(`/email/${accountId}`)
  },
}

// Database API
export const databaseAPI = {
  getDatabases: async (domainId: string): Promise<Database[]> => {
    const response: AxiosResponse<Database[]> = await api.get(`/domains/${domainId}/databases`)
    return response.data
  },

  createDatabase: async (domainId: string, data: {
    name: string
    type: string
  }): Promise<Database> => {
    const response: AxiosResponse<Database> = await api.post(`/domains/${domainId}/databases`, data)
    return response.data
  },

  deleteDatabase: async (databaseId: string): Promise<void> => {
    await api.delete(`/databases/${databaseId}`)
  },
}

// System API
export const systemAPI = {
  getStats: async (): Promise<{
    cpu: number
    memory: { used: number; total: number }
    disk: { used: number; total: number }
    uptime: number
  }> => {
    const response = await api.get('/system/stats')
    return response.data
  },

  getServices: async (): Promise<Array<{
    name: string
    status: string
    uptime: number
  }>> => {
    const response = await api.get('/system/services')
    return response.data
  },
}

export default api
