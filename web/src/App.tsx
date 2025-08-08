import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './stores/authStore';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Projects from './pages/Projects';
import Pipelines from './pages/Pipelines';
import Deployments from './pages/Deployments';
import SSHKeys from './pages/SSHKeys';
import Users from './pages/Users';
import Settings from './pages/Settings';

function App() {
  const { isAuthenticated } = useAuthStore();

  return (
    <Router>
      <div className="min-h-screen bg-background">
        <Routes>
          <Route 
            path="/login" 
            element={!isAuthenticated ? <Login /> : <Navigate to="/dashboard" />} 
          />
          <Route
            path="/*"
            element={
              isAuthenticated ? (
                <Layout>
                  <Routes>
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/projects/*" element={<Projects />} />
                    <Route path="/pipelines/*" element={<Pipelines />} />
                    <Route path="/deployments/*" element={<Deployments />} />
                    <Route path="/ssh-keys" element={<SSHKeys />} />
                    <Route path="/users" element={<Users />} />
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/" element={<Navigate to="/dashboard" />} />
                  </Routes>
                </Layout>
              ) : (
                <Navigate to="/login" />
              )
            }
          />
        </Routes>
      </div>
    </Router>
  );
}

export default App;