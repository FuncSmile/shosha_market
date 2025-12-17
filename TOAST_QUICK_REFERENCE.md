# Sonner Toast Implementation - Quick Reference

## How to Use Toast Notifications

### Basic Usage in Components

```typescript
import { useToast } from '../composables/useToast'

export default {
  setup() {
    const { success, error, warning, info, message } = useToast()
    
    return {
      success, error, warning, info, message
    }
  }
}
```

Or with Vue 3 `<script setup>`:

```vue
<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { success, error, warning, info, message } = useToast()

async function saveData() {
  try {
    await api.save(data)
    success('✓ Data saved successfully!')
  } catch (err) {
    error(`Error: ${err.message}`)
  }
}
</script>
```

## Toast Methods

### `success(message: string)` ✅
- **Duration**: 3 seconds
- **Color**: Green
- **Use for**: Successful operations (save, create, delete, export)

```typescript
success('✓ Product added successfully!')
```

### `error(message: string)` ❌
- **Duration**: 4 seconds
- **Color**: Red
- **Use for**: Errors and failures

```typescript
error('Failed to save: Invalid product name')
```

### `warning(message: string)` ⚠️
- **Duration**: 3 seconds
- **Color**: Yellow/Orange
- **Use for**: Validation warnings, alerts, non-critical issues

```typescript
warning('Please fill in all required fields')
```

### `info(message: string)` ℹ️
- **Duration**: 3 seconds
- **Color**: Blue
- **Use for**: Information messages

```typescript
info('Syncing data in background...')
```

### `message(msg: string)` 
- **Duration**: 3 seconds
- **Color**: Default (gray)
- **Use for**: General messages

```typescript
message('Operation in progress')
```

### `promise<T>(promise, messages)` ⏳
- **Type**: Promise wrapper
- **Use for**: Long-running operations with loading state

```typescript
const { promise } = useToast()

promise(
  api.createSale(saleData),
  {
    loading: 'Processing transaction...',
    success: '✓ Transaction saved!',
    error: 'Failed to save transaction'
  }
)
```

## Panel-Specific Toast Examples

### ProductPanel
```typescript
// When adding products
success(`✓ Berhasil menambahkan ${newRows.length} baris`)

// When duplicate detected
error(`Duplikat nama barang terdeteksi. ${parts.join(' | ')}`)

// When no data to save
warning('Tidak ada baris valid untuk disimpan')
```

### SalesPanel
```typescript
// Validation
warning('Lengkapi cabang dan pilih minimal 1 barang!')

// Success
success('Transaksi berhasil disimpan!')

// Error
error((err as Error).message)
```

### BranchPanel
```typescript
// Duplicate check
error(`Nama cabang sudah ada: ${dup.name}`)

// Save success
success('Berhasil menambah cabang')
```

### ReportsPanel
```typescript
// Export success
success('Laporan penjualan diunduh.')

// Export error
error((err as Error).message)
```

## Theme Customization

The Toaster is configured with:
```typescript
<Toaster theme="dark" position="top-right" />
```

### Available Themes
- `dark` (current - dark background with light text)
- `light` (light background with dark text)
- `system` (follows OS preference)

### Position Options
- `top-right` (current)
- `top-center`
- `top-left`
- `bottom-right`
- `bottom-center`
- `bottom-left`

## CSS Variables (Dark Theme)

The dark theme uses these CSS variables from `style.css`:

```css
:root[class~="dark"] {
  --background: 20 14.3% 4.1%;          /* Very dark gray */
  --foreground: 60 9.1% 97.8%;          /* Off-white */
  --primary: 20.5 90.2% 48.2%;          /* Orange */
  --secondary: 12 6.5% 15.1%;           /* Dark gray-brown */
  --muted: 12 6.5% 15.1%;               /* Dark gray-brown */
  --accent: 12 6.5% 15.1%;              /* Dark gray-brown */
}
```

## Important Notes

✅ **Do NOT:**
- Use `const message = ref('')` for alerts
- Display errors in `message.value` template bindings
- Create custom alert/notification elements

✅ **DO:**
- Import and use the `useToast()` composable
- Call appropriate toast method (success, error, warning, etc.)
- Let Sonner handle the UI and positioning
- Use consistent messages across similar operations

## Migration Checklist for New Components

- [ ] Import `useToast` from composables
- [ ] Destructure toast methods: `const { success, error, warning } = useToast()`
- [ ] Remove any `const message = ref('')` declarations
- [ ] Replace `message.value = 'msg'` with `success/error/warning('msg')`
- [ ] Remove message display elements from template
- [ ] Test all success and error paths
- [ ] Verify build passes: `npm run build`

## Files Modified

- ✅ `renderer/src/main.ts` - Toaster registered globally
- ✅ `renderer/src/App.vue` - Toaster component added to template
- ✅ `renderer/src/composables/useToast.ts` - Toast utility (NEW)
- ✅ `renderer/src/components/ProductPanel.vue`
- ✅ `renderer/src/components/BranchPanel.vue`
- ✅ `renderer/src/components/SalesPanel.vue`
- ✅ `renderer/src/components/DashboardPanel.vue`
- ✅ `renderer/src/components/StockOpnamePanel.vue`
- ✅ `renderer/src/components/ReportsPanel.vue`

## Troubleshooting

**Toasts not appearing?**
- Ensure `<Toaster theme="dark" position="top-right" />` is in App.vue template
- Check browser console for errors
- Verify Sonner package is installed: `npm ls vue-sonner`

**Wrong theme?**
- Check `theme` prop in `<Toaster>` component
- Verify CSS variables are defined in `style.css`
- Try `theme="light"` or `theme="system"`

**Type errors?**
- Verify import path is correct: `../composables/useToast`
- Check TypeScript version: `npm list typescript`
- Run type check: `npm run type-check` or `vue-tsc --noEmit`

---

For more info: See `SONNER_TOAST_IMPLEMENTATION.md`
