import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { IconChart, IconEye, IconLock } from '../components/Icons';
import './AuthPages.css';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await login({ email, password });
      const from = location.state?.from?.pathname || '/dashboard';
      navigate(from, { replace: true });
    } catch (err) {
      if (err.fields && err.fields.length > 0) {
        setError(err.fields.map(f => `${f.field}: ${f.message}`).join('. '));
      } else {
        setError(err.error || 'Invalid credentials. Please verify and try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page-wrapper">
      <div className="auth-ambient-glow" aria-hidden="true" />
      
      <div className="auth-container container">
        <div className="auth-card glass-panel-elevated animate-fade-in">
          <div className="auth-header">
            <div className="auth-brand-badge">
              <IconChart size={20} />
            </div>
            <h1 className="auth-title">Welcome back</h1>
            <p className="auth-subtitle">Sign in to manage and view your real-time polls</p>
          </div>

          {error && (
            <div className="auth-error-banner" role="alert">
              <span className="error-dot" />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="auth-form" noValidate>
            <div className="form-group">
              <label htmlFor="email" className="input-label">Email Address</label>
              <div className="input-wrapper">
                <input 
                  type="email" 
                  id="email" 
                  value={email} 
                  onChange={(e) => setEmail(e.target.value)}
                  required 
                  autoComplete="email"
                  placeholder="name@company.com"
                  disabled={loading}
                />
              </div>
            </div>

            <div className="form-group">
              <div className="label-row">
                <label htmlFor="password" className="input-label">Password</label>
              </div>
              <div className="input-wrapper password-wrapper">
                <input 
                  type={showPassword ? 'text' : 'password'} 
                  id="password" 
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)}
                  required 
                  autoComplete="current-password"
                  placeholder="••••••••••••"
                  disabled={loading}
                />
                <button 
                  type="button" 
                  className="password-toggle-btn"
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  <IconEye size={16} />
                </button>
              </div>
            </div>

            <button 
              type="submit" 
              className="btn btn-primary auth-submit-btn" 
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="button-spinner"></span>
                  <span>Signing in...</span>
                </>
              ) : (
                <span>Sign In to LivePoll</span>
              )}
            </button>
          </form>

          <div className="auth-footer">
            <span>Don't have an account?</span>{' '}
            <Link to="/signup" className="auth-link">Create an account</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
