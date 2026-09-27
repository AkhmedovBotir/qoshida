import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { AppLayout } from './layouts/AppLayout'
import { ActivityTypesPage } from './pages/ActivityTypesPage'
import { AdminsPage } from './pages/AdminsPage'
import { CategoriesPage } from './pages/CategoriesPage'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { ManagersPage } from './pages/ManagersPage'
import { RegionsPage } from './pages/RegionsPage'
import { ShopDirectorsPage } from './pages/ShopDirectorsPage'
import { KontragentsPage } from './pages/KontragentsPage'
import { LocalShopsPage } from './pages/LocalShopsPage'
import { ServiceProvidersPage } from './pages/ServiceProvidersPage'
import { IdentificationsPage } from './pages/IdentificationsPage'
import { ServicesPage } from './pages/ServicesPage'
import { SellersPage } from './pages/SellersPage'
import { DeliveriesPage } from './pages/DeliveriesPage'
import { ProductsPage } from './pages/ProductsPage'
import { ShopTemplatesPage } from './pages/ShopTemplatesPage'
import { ShopProductsPage } from './pages/ShopProductsPage'
import { ShopIncomingsPage } from './pages/ShopIncomingsPage'
import { SettingsPage } from './pages/SettingsPage'
import { SnackbarHost } from './components/ui/SnackbarHost'

export default function App() {
  return (
    <AuthProvider>
      <SnackbarHost />
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<AppLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="admins" element={<AdminsPage />} />
              <Route path="regions" element={<RegionsPage />} />
              <Route path="categories" element={<CategoriesPage />} />
              <Route path="activity-types" element={<ActivityTypesPage />} />
              <Route path="managers" element={<ManagersPage />} />
              <Route path="shop-directors" element={<ShopDirectorsPage />} />
              <Route path="kontragents" element={<KontragentsPage />} />
              <Route path="products" element={<ProductsPage />} />
              <Route path="shop-templates" element={<ShopTemplatesPage />} />
              <Route path="shop-products" element={<ShopProductsPage />} />
              <Route path="shop-incomings" element={<ShopIncomingsPage />} />
              <Route path="local-shops" element={<LocalShopsPage />} />
              <Route path="sellers" element={<SellersPage />} />
              <Route path="deliveries" element={<DeliveriesPage />} />
              <Route path="service-providers" element={<ServiceProvidersPage />} />
              <Route path="identifications" element={<IdentificationsPage />} />
              <Route path="services" element={<ServicesPage />} />
              <Route path="settings" element={<SettingsPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
