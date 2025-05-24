import { Routes, Route, Navigate } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { RootState } from './store'
import Layout from './components/Layout'
import LoginPage from './pages/auth/LoginPage'
import DashboardPage from './pages/DashboardPage'
import DomainsPage from './pages/domains/DomainsPage'
import EmailPage from './pages/email/EmailPage'
import DatabasesPage from './pages/databases/DatabasesPage'
import FilesPage from './pages/files/FilesPage'
import SystemPage from './pages/system/SystemPage'
import SettingsPage from './pages/settings/SettingsPage'

function App() {
  const { isAuthenticated } = useSelector((state: RootState) => state.auth)

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/domains/*" element={<DomainsPage />} />
        <Route path="/email/*" element={<EmailPage />} />
        <Route path="/databases/*" element={<DatabasesPage />} />
        <Route path="/files/*" element={<FilesPage />} />
        <Route path="/system/*" element={<SystemPage />} />
        <Route path="/settings/*" element={<SettingsPage />} />
        <Route path="/login" element={<Navigate to="/dashboard" replace />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Layout>
  )
}

export default App
