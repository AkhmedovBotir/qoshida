import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { IdentificationGate } from './components/auth/IdentificationGate'
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { ProviderLayout } from './layouts/ProviderLayout'
import { HomePage } from './pages/HomePage'
import { IdentificationPage } from './pages/IdentificationPage'
import { ServicesPage } from './pages/ServicesPage'
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
            <Route element={<ProviderLayout />}>
              <Route element={<IdentificationGate />}>
                <Route index element={<HomePage />} />
                <Route path="identification" element={<IdentificationPage />} />
                <Route path="services" element={<ServicesPage />} />
              </Route>
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
