import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from './hooks/AuthContext';
import { ToastProvider } from '../components/Toast';
import ProtectedRoute from '../components/ProtectedRoute';
import Layout from '../components/Layout';
import LandingPage from './pages/landing/page';
import Login from './pages/login/page';
import Dashboard from './pages/dashboard/page';
import CaseManagement from './pages/cases/page';
import Configuration from './pages/settings/page';
import BlockchainExplorer from './pages/explorer/page';
import Reports from './pages/reports/page';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ToastProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/" element={<LandingPage />} />
              <Route path="/login" element={<Login />} />

              <Route element={<ProtectedRoute />}>
                <Route element={<Layout />}>
                  <Route path="/dashboard" element={<Dashboard />} />
                  <Route path="/cases" element={<CaseManagement />} />
                  <Route path="/explorer" element={<BlockchainExplorer />} />
                  <Route path="/reports" element={<Reports />} />
                  <Route path="/settings" element={<Configuration />} />
                </Route>
              </Route>

              {/* Public SaaS Auditor Portal Routes */}
              <Route element={<Layout />}>
                <Route path="/public/:companyId/dashboard" element={<Dashboard />} />
                <Route path="/public/:companyId/cases" element={<CaseManagement />} />
                <Route path="/public/:companyId/explorer" element={<BlockchainExplorer />} />
              </Route>

              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </BrowserRouter>
        </ToastProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;

