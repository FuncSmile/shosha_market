import type { User, RegisterDto, PaginatedResponse } from '~/types'

export interface UpdateUserDto {
  name?: string
  email?: string
  role?: string
  userCode?: string
  isActive?: boolean
  password?: string
}

export const useUsersApi = () => {
  const api = useApi()

  return {
    // Get all users with filters
    getAllUsers: (filters?: { role?: string; userCode?: string; isActive?: boolean }) => {
      const params = new URLSearchParams()
      if (filters?.role) params.append('role', filters.role)
      if (filters?.userCode) params.append('userCode', filters.userCode)
      if (filters?.isActive !== undefined) params.append('isActive', String(filters.isActive))
      const query = params.toString() ? `?${params.toString()}` : ''
      return api.get<PaginatedResponse<User> | User[]>(`/users${query}`)
    },

    // Get user by ID
    getUserById: (id: string) =>
      api.get<User>(`/users/${id}`),

    // Create new user
    createUser: (data: RegisterDto) =>
      api.post<User>('/users', data),

    // Update user
    updateUser: (id: string, data: UpdateUserDto) =>
      api.patch<User>(`/users/${id}`, data),

    // Soft delete user
    deleteUser: (id: string) =>
      api.delete(`/users/${id}`),

    // Hard delete user
    hardDeleteUser: (id: string) =>
      api.delete(`/users/${id}/hard`),
  }
}
