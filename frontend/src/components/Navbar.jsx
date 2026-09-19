import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { IconChart, IconPlus, IconLogout, IconUser } from './Icons';
import './Navbar.css';

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    setMobileMenuOpen(false);
    navigate('/');
  };

  const isCurrent = (path) => location.pathname === path;

  return (
    <header className="navbar-wrapper">
      <nav className="navbar container">
        <Link to="/" className="navbar-brand" onClick={() => setMobileMenuOpen(false)}>
          <div className="brand-logo-glow">
            <IconChart size={20} className="brand-icon" />
          </div>
          <span className="brand-name">
            Live<span className="gradient-text">Poll</span>
          </span>
          <span className="brand-tag">PRO</span>
        </Link>

        {/* Desktop Navigation */}
        <div className="navbar-desktop">
          {user ? (
            <div className="nav-group">
              <Link 
                to="/dashboard" 
                className={`nav-link ${isCurrent('/dashboard') ? 'active' : ''}`}
              >
                Dashboard
              </Link>
              <Link to="/create" className="btn btn-primary btn-sm nav-cta">
                <IconPlus size={15} />
                <span>Create Poll</span>
              </Link>
              
              <div className="user-profile-menu">
                <div className="user-avatar" title={user.email || user.name}>
                  {user.name ? user.name.charAt(0).toUpperCase() : <IconUser size={14} />}
                </div>
                <span className="user-display-name">{user.name}</span>
                <button 
                  onClick={handleLogout} 
                  className="logout-btn" 
                  title="Sign out"
                  aria-label="Sign out"
                >
                  <IconLogout size={16} />
                </button>
              </div>
            </div>
          ) : (
            <div className="nav-group">
              <Link 
                to="/login" 
                className={`nav-link ${isCurrent('/login') ? 'active' : ''}`}
              >
                Sign In
              </Link>
              <Link to="/signup" className="btn btn-primary btn-sm nav-cta">
                <span>Get Started Free</span>
              </Link>
            </div>
          )}
        </div>

        {/* Mobile Menu Toggle */}
        <button 
          className="mobile-toggle" 
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          aria-label="Toggle navigation menu"
        >
          <span className={`hamburger-bar ${mobileMenuOpen ? 'open' : ''}`}></span>
        </button>
      </nav>

      {/* Mobile Drawer */}
      {mobileMenuOpen && (
        <div className="mobile-drawer animate-fade-in">
          {user ? (
            <div className="mobile-nav-items">
              <div className="mobile-user-info">
                <div className="user-avatar">{user.name?.charAt(0).toUpperCase()}</div>
                <div>
                  <div className="user-display-name">{user.name}</div>
                  <div className="mobile-user-email">{user.email}</div>
                </div>
              </div>
              <Link 
                to="/dashboard" 
                className="mobile-nav-link"
                onClick={() => setMobileMenuOpen(false)}
              >
                Dashboard
              </Link>
              <Link 
                to="/create" 
                className="btn btn-primary"
                onClick={() => setMobileMenuOpen(false)}
              >
                <IconPlus size={16} />
                <span>Create Poll</span>
              </Link>
              <button onClick={handleLogout} className="btn btn-secondary mobile-logout">
                <IconLogout size={16} />
                <span>Sign Out</span>
              </button>
            </div>
          ) : (
            <div className="mobile-nav-items">
              <Link 
                to="/login" 
                className="mobile-nav-link"
                onClick={() => setMobileMenuOpen(false)}
              >
                Sign In
              </Link>
              <Link 
                to="/signup" 
                className="btn btn-primary"
                onClick={() => setMobileMenuOpen(false)}
              >
                <span>Get Started Free</span>
              </Link>
            </div>
          )}
        </div>
      )}
    </header>
  );
}
