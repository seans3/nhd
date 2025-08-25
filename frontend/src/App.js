import React from 'react';
import { BrowserRouter as Router, Route, Routes, useLocation } from 'react-router-dom';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Financials from './pages/Financials';
import Navbar from './components/Navbar';

// A component to conditionally render the Navbar
const AppNavbar = () => {
  const location = useLocation();
  // Don't show Navbar on the login page
  if (location.pathname === '/login') {
    return null;
  }
  return <Navbar />;
};

function App() {
  return (
    <Router>
      <AppNavbar />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/financials" element={<Financials />} />
        <Route path="/" element={<Dashboard />} />
      </Routes>
    </Router>
  );
}

export default App;
