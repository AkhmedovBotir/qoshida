import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { DirectorLayout } from './layouts/DirectorLayout'
import { HomePage } from './pages/HomePage'
import { KontragentsPage } from './pages/KontragentsPage'
import { ProductsPage } from './pages/ProductsPage'
import { ShopTemplatesPage } from './pages/ShopTemplatesPage'
import { ShopProductsPage } from './pages/ShopProductsPage'
import { ShopIncomingsPage } from './pages/ShopIncomingsPage'
import { LocalShopsPage } from './pages/LocalShopsPage'
import { ServiceProvidersPage } from './pages/ServiceProvidersPage'
import { IdentificationsPage } from './pages/IdentificationsPage'
import { ServicesPage } from './pages/ServicesPage'
import { SellersPage } from './pages/SellersPage'
import { DeliveriesPage } from './pages/DeliveriesPage'
import { CategoriesPage } from './pages/CategoriesPage'
import { LoginPage } from './pages/LoginPage'
import { SnackbarHost } from './components/ui/SnackbarHost'

export default function App() {
  return (
    <AuthProvider>
      <SnackbarHost />
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<DirectorLayout />}>
              <Route index element={<HomePage />} />
              <Route path="kontragents" element={<KontragentsPage />} />
              <Route path="products" element={<ProductsPage />} />
              <Route path="shop-templates" element={<ShopTemplatesPage />} />
              <Route path="shop-products" element={<ShopProductsPage />} />
              <Route path="shop-incomings" element={<ShopIncomingsPage />} />
              <Route path="local-shops" element={<LocalShopsPage />} />
              <Route path="sellers" element={<SellersPage />} />
              <Route path="deliveries" element={<DeliveriesPage />} />
              <Route path="categories" element={<CategoriesPage />} />
              <Route path="service-providers" element={<ServiceProvidersPage />} />
              <Route path="identifications" element={<IdentificationsPage />} />
              <Route path="services" element={<ServicesPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
