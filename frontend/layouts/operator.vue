<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Sidebar untuk Desktop -->
    <aside class="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
      <div class="flex flex-col flex-grow bg-white border-r overflow-y-auto">
        <!-- Logo -->
        <div class="flex items-center flex-shrink-0 px-6 py-4 border-b">
          <span class="text-xl font-bold text-primary-600">ShoshaMart</span>
        </div>

        <!-- User Info -->
        <div class="px-6 py-4 border-b">
          <p class="text-sm font-semibold text-gray-900">{{ authStore.user?.name }}</p>
          <p class="text-xs text-gray-500">{{ getRoleLabel(authStore.user?.role!) }}</p>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 px-3 py-4 space-y-1">
          <NuxtLink
            v-for="item in navigationItems"
            :key="item.path"
            :to="item.path"
            class="group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors"
            :class="$route.path === item.path 
              ? 'bg-primary-50 text-primary-600' 
              : 'text-gray-700 hover:bg-gray-50 hover:text-primary-600'"
          >
            <Icon :name="item.icon" class="mr-3 h-5 w-5 flex-shrink-0" />
            {{ item.label }}
          </NuxtLink>
        </nav>

        <!-- Logout Button -->
        <div class="flex-shrink-0 border-t p-4">
          <button
            @click="handleLogout"
            class="w-full flex items-center px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50 rounded-lg transition-colors"
          >
            <Icon name="lucide:log-out" class="mr-3 h-5 w-5" />
            Keluar
          </button>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div class="lg:pl-64 flex flex-col min-h-screen">
      <!-- Top Header untuk Mobile -->
      <header class="lg:hidden sticky top-0 z-10 flex h-16 flex-shrink-0 bg-white border-b">
        <div class="flex flex-1 items-center justify-between px-4">
          <span class="text-lg font-bold text-primary-600">ShoshaMart</span>
          <div class="flex items-center space-x-2">
            <span class="text-sm font-medium text-gray-900">{{ authStore.user?.name }}</span>
            <button @click="handleLogout" class="p-2 text-gray-500 hover:text-red-600 rounded-full hover:bg-gray-100 transition-colors" aria-label="Keluar">
              <Icon name="lucide:log-out" class="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      <!-- Main Content -->
      <main class="flex-1 pb-20 lg:pb-0">
        <div class="px-4 py-6 sm:px-6 lg:px-8">
          <slot />
        </div>
      </main>
    </div>

    <!-- Bottom Navigation untuk Mobile -->
    <nav class="lg:hidden fixed bottom-0 left-0 right-0 bg-white border-t z-10">
      <div class="grid grid-cols-5 h-16">
        <NuxtLink
          v-for="item in mobileNavigationItems"
          :key="item.path"
          :to="item.path"
          class="flex flex-col items-center justify-center text-xs font-medium transition-colors"
          :class="$route.path === item.path 
            ? 'text-primary-600' 
            : 'text-gray-500 hover:text-primary-600'"
        >
          <Icon :name="item.icon" class="h-6 w-6 mb-1" />
          <span>{{ item.shortLabel || item.label }}</span>
        </NuxtLink>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { getRoleLabel } from '~/utils/helpers'

const authStore = useAuthStore()

const navigationItems = computed(() => [
  { 
    path: '/operator', 
    label: 'Dashboard', 
    shortLabel: 'Home',
    icon: 'lucide:layout-dashboard' 
  },
  { 
    path: '/operator/products', 
    label: 'Produk',
    icon: 'lucide:package' 
  },
  { 
    path: '/operator/orders', 
    label: 'Pesanan',
    icon: 'lucide:shopping-bag' 
  },
  { 
    path: '/operator/pricing', 
    label: 'Harga',
    icon: 'lucide:tag' 
  },
  { 
    path: '/operator/users', 
    label: 'Users',
    icon: 'lucide:users' 
  },
  { 
    path: '/operator/reports', 
    label: 'Laporan',
    icon: 'lucide:file-text' 
  },
])

// Mobile navigation (5 items utama)
const mobileNavigationItems = computed(() => [
  { 
    path: '/operator', 
    label: 'Dashboard', 
    shortLabel: 'Home',
    icon: 'lucide:layout-dashboard' 
  },
  { 
    path: '/operator/products', 
    label: 'Produk',
    icon: 'lucide:package' 
  },
  { 
    path: '/operator/orders', 
    label: 'Pesanan',
    icon: 'lucide:shopping-bag' 
  },
  { 
    path: '/operator/pricing', 
    label: 'Harga',
    icon: 'lucide:tag' 
  },
  { 
    path: '/operator/users', 
    label: 'Users',
    icon: 'lucide:users' 
  },
])

const handleLogout = async () => {
  await authStore.logout()
}
</script>
