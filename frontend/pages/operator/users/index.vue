<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Manajemen User</h1>
        <p class="mt-1 text-sm text-gray-500">Kelola semua pengguna sistem</p>
      </div>
      <button
        @click="showCreateModal = true"
        class="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
      >
        <Icon name="lucide:plus" class="w-5 h-5 mr-2" />
        Tambah User
      </button>
    </div>

    <!-- Filters -->
    <div class="bg-white rounded-lg shadow-sm p-4">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
          <select
            v-model="filters.role"
            @change="loadUsers"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          >
            <option value="">Semua Role</option>
            <option value="OPERATOR">Operator</option>
            <option value="ADMIN">Admin</option>
            <option value="PEMBELI">Pembeli</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">User Code</label>
          <input
            v-model="filters.userCode"
            @input="loadUsers"
            type="text"
            placeholder="Contoh: SHOSHA"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select
            v-model="filters.isActive"
            @change="loadUsers"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
          >
            <option value="">Semua Status</option>
            <option :value="true">Aktif</option>
            <option :value="false">Nonaktif</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Users Table -->
    <div class="bg-white rounded-lg shadow-sm overflow-hidden">
      <div v-if="loading" class="p-8 text-center">
        <Icon name="lucide:loader-2" class="w-8 h-8 animate-spin mx-auto text-gray-400" />
        <p class="mt-2 text-gray-500">Memuat data...</p>
      </div>

      <div v-else-if="users.length === 0" class="p-8 text-center">
        <Icon name="lucide:users" class="w-12 h-12 mx-auto text-gray-400" />
        <p class="mt-2 text-gray-500">Tidak ada user ditemukan</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Nama</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">User Code</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Aksi</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="user in users" :key="user.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900">{{ user.name }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-500">{{ user.email }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-purple-100 text-purple-800': user.role === 'OPERATOR',
                    'bg-blue-100 text-blue-800': user.role === 'ADMIN',
                    'bg-green-100 text-green-800': user.role === 'PEMBELI'
                  }"
                >
                  {{ getRoleLabel(user.role) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ user.userCode || '-' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="user.isActive ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                >
                  {{ user.isActive ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                <button
                  @click="editUser(user)"
                  class="text-primary-700 hover:text-primary-900 hover:bg-primary-50 rounded p-1 focus:ring-2 focus:ring-primary-500 focus:ring-offset-1"
                >
                  <Icon name="lucide:edit" class="w-5 h-5" />
                </button>
                <button
                  @click="confirmDelete(user)"
                  class="text-red-700 hover:text-red-900 hover:bg-red-50 rounded p-1 focus:ring-2 focus:ring-red-500 focus:ring-offset-1"
                >
                  <Icon name="lucide:trash-2" class="w-5 h-5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || showEditModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-lg max-w-md w-full p-6">
        <h3 class="text-lg font-semibold mb-4">
          {{ showEditModal ? 'Edit User' : 'Tambah User Baru' }}
        </h3>

        <form @submit.prevent="showEditModal ? handleUpdate() : handleCreate()" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Nama</label>
            <input
              v-model="formData.name"
              required
              type="text"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              v-model="formData.email"
              required
              type="email"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div v-if="!showEditModal">
            <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <input
              v-model="formData.password"
              required
              type="password"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
            <select
              v-model="formData.role"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            >
              <option value="OPERATOR">Operator</option>
              <option value="ADMIN">Admin</option>
              <option value="PEMBELI">Pembeli</option>
            </select>
          </div>

          <div v-if="formData.role !== 'OPERATOR'">
            <label class="block text-sm font-medium text-gray-700 mb-1">User Code</label>
            <input
              v-model="formData.userCode"
              type="text"
              placeholder="Contoh: SHOSHA, L24J"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div v-if="showEditModal" class="flex items-center">
            <input
              v-model="formData.isActive"
              type="checkbox"
              id="isActive"
              class="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
            />
            <label for="isActive" class="ml-2 block text-sm text-gray-700">
              User Aktif
            </label>
          </div>

          <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded-lg">
            <p class="text-sm text-red-600">{{ error }}</p>
          </div>

          <div class="flex gap-3 pt-4">
            <button
              type="button"
              @click="closeModal"
              class="flex-1 px-4 py-2 border-2 border-secondary-300 text-secondary-700 font-semibold rounded-lg hover:bg-secondary-50 hover:border-secondary-400 focus:ring-2 focus:ring-secondary-500 focus:ring-offset-2 shadow-sm bg-white"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="submitting"
              class="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
            >
              {{ submitting ? 'Menyimpan...' : 'Simpan' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { User, RegisterDto } from '~/types'
import { getRoleLabel } from '~/utils/helpers'

definePageMeta({
  layout: 'operator',
  middleware: 'role',
})

const usersApi = useUsersApi()

const toast = useToast()

const loading = ref(false)
const users = ref<User[]>([])
const showCreateModal = ref(false)
const showEditModal = ref(false)
const submitting = ref(false)
const error = ref('')

const filters = reactive({
  role: '',
  userCode: '',
  isActive: '' as '' | boolean,
})

const formData = reactive<RegisterDto & { id?: string; isActive?: boolean }>({
  name: '',
  email: '',
  password: '',
  role: 'PEMBELI' as any,
  userCode: '',
})

const loadUsers = async () => {
  loading.value = true
  try {
    const filterParams: any = {}
    if (filters.role) filterParams.role = filters.role
    if (filters.userCode) filterParams.userCode = filters.userCode
    if (filters.isActive !== '') filterParams.isActive = filters.isActive

    const response = await usersApi.getAllUsers(filterParams)
    // Handle paginated response
    if (response && 'data' in response && Array.isArray(response.data)) {
      users.value = response.data
    } else if (Array.isArray(response)) {
      // Fallback for non-paginated response
      users.value = response
    } else {
      users.value = []
    }
  } catch (err: any) {
    console.error('Failed to load users:', err)
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  submitting.value = true
  error.value = ''
  try {
    await usersApi.createUser({
      name: formData.name,
      email: formData.email,
      password: formData.password,
      role: formData.role,
      userCode: formData.role === 'OPERATOR' ? undefined : formData.userCode,
    })
    closeModal()
    await loadUsers()
    toast.success('User berhasil dibuat', 'Berhasil')
  } catch (err: any) {
    error.value = err.message || 'Gagal membuat user'
  } finally {
    submitting.value = false
  }
}

const editUser = (user: User) => {
  formData.id = user.id
  formData.name = user.name
  formData.email = user.email
  formData.role = user.role as any
  formData.userCode = user.userCode || ''
  formData.isActive = user.isActive
  showEditModal.value = true
}

const handleUpdate = async () => {
  if (!formData.id) return
  
  submitting.value = true
  error.value = ''
  try {
    await usersApi.updateUser(formData.id, {
      name: formData.name,
      email: formData.email,
      role: formData.role,
      userCode: formData.role === 'OPERATOR' ? undefined : formData.userCode,
      isActive: formData.isActive,
    })
    closeModal()
    await loadUsers()
    toast.success('User berhasil diupdate', 'Berhasil')
  } catch (err: any) {
    error.value = err.message || 'Gagal update user'
  } finally {
    submitting.value = false
  }
}

const confirmDelete = async (user: User) => {
  if (!confirm(`Hapus user ${user.name}?`)) return
  
  try {
    await usersApi.deleteUser(user.id)
    await loadUsers()
    toast.success('User berhasil dihapus', 'Berhasil')
  } catch (err: any) {
    toast.error(err.message || 'Gagal menghapus user', 'Gagal')
  }
}

const closeModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  error.value = ''
  Object.assign(formData, {
    id: undefined,
    name: '',
    email: '',
    password: '',
    role: 'PEMBELI',
    userCode: '',
    isActive: true,
  })
}

onMounted(() => {
  loadUsers()
})
</script>
