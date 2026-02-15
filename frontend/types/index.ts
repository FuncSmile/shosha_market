// ========================================
// USER TYPES
// ========================================

export enum UserRole {
  OPERATOR = 'OPERATOR',
  ADMIN = 'ADMIN',
  PEMBELI = 'PEMBELI',
}

export interface User {
  id: string
  email: string
  name: string
  role: UserRole
  userCode: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface LoginDto {
  email: string
  password: string
}

export interface RegisterDto {
  email: string
  password: string
  name: string
  role: UserRole
  userCode?: string
}

export interface AuthResponse {
  user: User
  accessToken: string
}

// ========================================
// PRODUCT TYPES
// ========================================

export interface Product {
  id: string
  name: string
  description: string | null
  category: string
  imageUrl: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
  variants?: ProductVariant[]
}

export interface ProductVariant {
  id: string
  productId: string
  sku: string
  name: string
  attributes: Record<string, any> | null
  stock: number
  isActive: boolean
  createdAt: string
  updatedAt: string
  product?: Product
  pricing?: Pricing[]
}

export interface CreateProductDto {
  name: string
  description?: string
  category: string
  imageUrl?: string
  isActive?: boolean
}

export interface UpdateProductDto {
  name?: string
  description?: string
  category?: string
  imageUrl?: string
  isActive?: boolean
}

export interface CreateVariantDto {
  productId: string
  sku: string
  name: string
  attributes?: Record<string, any>
  stock: number
}

export interface UpdateVariantDto {
  sku?: string
  name?: string
  attributes?: Record<string, any>
  stock?: number
  isActive?: boolean
}

// ========================================
// PRICING TYPES
// ========================================

export interface Pricing {
  id: string
  variantId: string
  userCode: string
  price: number
  createdAt: string
  updatedAt: string
  variant?: ProductVariant
}

export interface SetPricingDto {
  variantId: string
  userCode: string
  price: number
}

export interface BulkPricingDto {
  items: SetPricingDto[]
}

export interface PricingWithProduct extends Pricing {
  variant: ProductVariant & {
    product: Product
  }
}

// ========================================
// ORDER TYPES
// ========================================

export enum OrderStatus {
  PENDING = 'PENDING',
  APPROVED_BY_ADMIN = 'APPROVED_BY_ADMIN',
  DISETUJUI = 'DISETUJUI',
  DIKEMAS = 'DIKEMAS',
  SEGERA_DIKIRIM = 'SEGERA_DIKIRIM',
  SELESAI = 'SELESAI',
  DITOLAK = 'DITOLAK',
}

export interface Order {
  id: string
  orderNumber: string
  buyerId: string
  buyerUserCode: string
  totalAmount: number
  status: OrderStatus
  approvedByAdmin: boolean
  approvedByOperator: boolean
  approvedAt: string | null
  invoicePrintedAt: string | null
  createdAt: string
  updatedAt: string
  buyer?: User
  items?: OrderItem[]
  comments?: Comment[]
}

export interface OrderItem {
  id: string
  orderId: string
  variantId: string
  productName: string
  variantName: string
  sku: string
  pricePerUnit: number
  quantity: number
  subtotal: number
  createdAt: string
  updatedAt: string
  variant?: ProductVariant
}

export interface CreateOrderDto {
  items: {
    variantId: string
    quantity: number
  }[]
}

export interface UpdateOrderStatusDto {
  status: OrderStatus
  comment?: string
}

export interface EditOrderDto {
  items: {
    variantId: string
    quantity: number
  }[]
  comment: string
}

export interface AuditLog {
  id: string
  userId: string
  userName: string
  userRole: UserRole
  orderId: string
  action: string
  fieldChanged: string | null
  oldValue: string | null
  newValue: string | null
  notes: string | null
  createdAt: string
}

// ========================================
// COMMENT TYPES
// ========================================

export interface Comment {
  id: string
  orderId: string
  userId: string
  content: string
  isRequired: boolean
  isRead: boolean
  createdAt: string
  updatedAt: string
  user?: User
  order?: Order
}

export interface CreateCommentDto {
  orderId: string
  content: string
  isRequired?: boolean
}

// ========================================
// REPORT TYPES
// ========================================

export interface SalesReportParams {
  startDate: string
  endDate: string
  userCode?: string
}

export interface SalesReportData {
  totalOrders: number
  totalRevenue: number
  ordersByStatus: Record<OrderStatus, number>
  topProducts: {
    productName: string
    variantName: string
    totalQuantity: number
    totalRevenue: number
  }[]
}

// ========================================
// API RESPONSE TYPES
// ========================================

export interface ApiResponse<T = any> {
  data?: T
  message?: string
  error?: string
}

export interface PaginatedResponse<T = any> {
  data: T[]
  meta: {
    page: number
    limit: number
    total: number
    totalPages: number
    hasNextPage: boolean
    hasPreviousPage: boolean
  }
}

// ========================================
// DASHBOARD TYPES
// ========================================

export * from './dashboard'
