import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { DistrictOnly, ProtectedRoute, RegionOnly } from './components/auth/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { ManagerLayout } from './layouts/ManagerLayout'
import { DistrictHomePage } from './pages/district/DistrictHomePage'
import { LoginPage } from './pages/LoginPage'
import { ShopDirectorsPage } from './pages/ShopDirectorsPage'
import { KontragentsPage } from './pages/KontragentsPage'
import { LocalShopsPage } from './pages/LocalShopsPage'
import { ServiceProvidersPage } from './pages/ServiceProvidersPage'
import { IdentificationsPage } from './pages/IdentificationsPage'
import { ServicesPage } from './pages/ServicesPage'
import { SellersPage } from './pages/SellersPage'
import { DeliveriesPage } from './pages/DeliveriesPage'
import { CategoriesPage } from './pages/CategoriesPage'
import { ProductsPage } from './pages/ProductsPage'
import { ShopTemplatesPage } from './pages/ShopTemplatesPage'
import { ShopProductsPage } from './pages/ShopProductsPage'
import { ShopIncomingsPage } from './pages/ShopIncomingsPage'
import { DistrictManagersPage } from './pages/region/DistrictManagersPage'
import { RegionHomePage } from './pages/region/RegionHomePage'
import { SnackbarHost } from './components/ui/SnackbarHost'

export default function App() {
  return (
    <AuthProvider>
      <SnackbarHost />
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<RegionOnly />}>
              <Route element={<ManagerLayout />}>
                <Route path="/region" element={<RegionHomePage />} />
                <Route path="/region/district-managers" element={<DistrictManagersPage />} />
                <Route path="/region/shop-directors" element={<ShopDirectorsPage />} />
                <Route path="/region/kontragents" element={<KontragentsPage />} />
                <Route path="/region/products" element={<ProductsPage />} />
                <Route path="/region/shop-templates" element={<ShopTemplatesPage />} />
                <Route path="/region/shop-products" element={<ShopProductsPage />} />
                <Route path="/region/shop-incomings" element={<ShopIncomingsPage />} />
                <Route path="/region/local-shops" element={<LocalShopsPage />} />
                <Route path="/region/sellers" element={<SellersPage />} />
                <Route path="/region/deliveries" element={<DeliveriesPage />} />
                <Route path="/region/categories" element={<CategoriesPage />} />
                <Route path="/region/service-providers" element={<ServiceProvidersPage />} />
                <Route path="/region/identifications" element={<IdentificationsPage />} />
                <Route path="/region/services" element={<ServicesPage />} />
              </Route>
            </Route>
            <Route element={<DistrictOnly />}>
              <Route element={<ManagerLayout />}>
                <Route path="/district" element={<DistrictHomePage />} />
                <Route path="/district/shop-directors" element={<ShopDirectorsPage />} />
                <Route path="/district/kontragents" element={<KontragentsPage />} />
                <Route path="/district/products" element={<ProductsPage />} />
                <Route path="/district/shop-templates" element={<ShopTemplatesPage />} />
                <Route path="/district/shop-products" element={<ShopProductsPage />} />
                <Route path="/district/shop-incomings" element={<ShopIncomingsPage />} />
                <Route path="/district/local-shops" element={<LocalShopsPage />} />
                <Route path="/district/sellers" element={<SellersPage />} />
                <Route path="/district/deliveries" element={<DeliveriesPage />} />
                <Route path="/district/categories" element={<CategoriesPage />} />
                <Route path="/district/service-providers" element={<ServiceProvidersPage />} />
                <Route path="/district/identifications" element={<IdentificationsPage />} />
                <Route path="/district/services" element={<ServicesPage />} />
              </Route>
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
