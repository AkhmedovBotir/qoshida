import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { SellerLayout } from './layouts/SellerLayout'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { ShopProductsPage } from './pages/ShopProductsPage'
import { ShopIncomingsPage } from './pages/ShopIncomingsPage'
import { SnackbarHost } from './components/ui/SnackbarHost'

export default function App() {
  return (
    <AuthProvider>
      <SnackbarHost />
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<SellerLayout />}>
              <Route index element={<HomePage />} />
              <Route path="shop-products" element={<ShopProductsPage />} />
              <Route path="shop-incomings" element={<ShopIncomingsPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
