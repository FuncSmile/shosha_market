export interface Product {
  id: string
  name: string
  sku: string
  stock: number
  price: number
  synced: boolean
  created_at: string
  updated_at: string
}

export interface Branch {
  id: string
  code?: string
  name: string
  address?: string
  phone?: string
  synced?: boolean
  created_at?: string
  updated_at?: string
}

export interface SaleItem {
  id: string
  sale_id: string
  product_id: string
  qty: number
  price: number
  subtotal: number
  created_at: string
  updated_at: string
}

export interface Sale {
  id: string
  branch_id: string
  receipt_no: string
  total: number
  synced: boolean
  created_at: string
  updated_at: string
  items: SaleItem[]
}

export interface StockOpname {
  id: string
  branch_id: string
  note: string
  synced: boolean
  created_at: string
  updated_at: string
}

export interface SalesAnalytics {
  start: string
  end: string
  totalRevenue: number
  totalOrders: number
  totalItems: number
  perDay: { day: string; orders: number; items: number; revenue: number }[]
}

export interface SyncSummary {
  queuedChanges: number
  lastSyncAt: string | null
  dbPath: string
  status: string
  lastError?: string
}

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://127.0.0.1:8080/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
    ...options,
  });
  if (!res.ok) {
    const msg = await res.text();
    throw new Error(msg || res.statusText);
  }
  const contentType = res.headers.get('content-type') ?? '';
  if (contentType.includes('application/json')) {
    return res.json() as Promise<T>;
  }
  return undefined as T;
}

export const api = {
  listProducts: () => request<Product[]>('/products'),
  createProduct: (payload: Partial<Product>) =>
    request<Product>('/products', { method: 'POST', body: JSON.stringify(payload) }),
  updateProduct: (id: string, payload: Partial<Product>) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  deleteProduct: (id: string) => request<void>(`/products/${id}`, { method: 'DELETE' }),

  listBranches: () => request<Branch[]>('/branches'),
  createBranch: (payload: Partial<Branch>) =>
    request<Branch>('/branches', { method: 'POST', body: JSON.stringify(payload) }),
  updateBranch: (id: string, payload: Partial<Branch>) =>
    request<Branch>(`/branches/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  deleteBranch: (id: string) => request<void>(`/branches/${id}`, { method: 'DELETE' }),

  createSale: (payload: { branch_id: string; receipt_no: string; items: { product_id: string; qty: number; price: number }[] }) =>
    request<Sale>('/sales', { method: 'POST', body: JSON.stringify(payload) }),

  createStockOpname: (payload: { branch_id: string; note: string; items: { product_id: string; qty_system: number; qty_physical: number }[] }) =>
    request('/stock-opname', { method: 'POST', body: JSON.stringify(payload) }),

  downloadSalesReport: async (start: string, end: string) => {
    const res = await fetch(`${API_BASE}/reports/sales?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`);
    if (!res.ok) throw new Error(await res.text());
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    return url;
  },

  downloadOpnameReport: async (id: string) => {
    const res = await fetch(`${API_BASE}/stock-opname/${id}/report`);
    if (!res.ok) throw new Error(await res.text());
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    return url;
  },

  salesAnalytics: (start?: string, end?: string) => {
    const qs = new URLSearchParams();
    if (start) qs.append('start', start);
    if (end) qs.append('end', end);
    const suffix = qs.toString() ? `?${qs.toString()}` : '';
    return request<SalesAnalytics>(`/analytics/sales${suffix}`);
  },

  syncSummary: () => request<SyncSummary>('/sync/summary'),
  syncRun: () => request<{ status: string }>('/sync/run', { method: 'POST' }),
};
