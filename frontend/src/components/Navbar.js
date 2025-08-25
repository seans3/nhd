import React from 'react';
import { Link } from 'react-router-dom';

function Navbar() {
  const navStyle = {
    background: '#333',
    color: '#fff',
    padding: '10px',
    display: 'flex',
    justifyContent: 'space-around',
  };

  const linkStyle = {
    color: '#fff',
    textDecoration: 'none',
  };

  return (
    <nav style={navStyle}>
      <Link to="/dashboard" style={linkStyle}>Dashboard</Link>
      <Link to="/financials" style={linkStyle}>Financials</Link>
    </nav>
  );
}

export default Navbar;
