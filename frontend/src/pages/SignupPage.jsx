import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { IconChart, IconEye, IconCheck } from '../components/Icons';
import './AuthPages.css';

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { signup } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (name.trim().length === 0) {
      return setError('Please enter your full name');
    }

    if (password.length < 8) {
      return setError('Password must be at least 8 characters');
    }

    setLoading(true);

    try {
      await signup({ name: name.trim(), email: email.trim(), password });
      navigate('/dashboard');
    } catch (err) {
      if (err.fields && err.fields.length > 0) {
        setError(err.fields.map(f => `${f.field}: ${f.message}`).join('. '));
      } else {
        setError(err.error || 'Failed to create account. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const hasLength = password.length >= 8;
  const hasStrongLength = password.length >= 8;
  const hasUpper = /[A-Z]/.test(password);
  const hasNumber = /[0-9]/.test(password);

  const calculateStrength = () => {
    if (password.length === 0) return 0;
    let strength = 0;
    if (hasLength) strength += 25;
    if (hasStrongLength) strength += 25;
    if (hasUpper) strength += 25;
    if (hasNumber) strength += 25;
    return strength;
  };

  const strength = calculateStrength();
  let strengthLabel = 'Weak';
  let strengthClass = 'strength-weak';
  if (strength >= 50 && strength < 75) {
    strengthLabel = 'Fair';
    strengthClass = 'strength-fair';
  } else if (strength >= 75) {
    strengthLabel = 'Strong';
    strengthClass = 'strength-strong';
  }

  return (
    <div className="auth-page-wrapper">
      <div className="auth-ambient-glow" aria-hidden="true" />

      <div className="auth-container container">
        <div className="auth-card glass-panel-elevated animate-fade-in">
          <div className="auth-header">
            <div className="auth-brand-badge">
              <IconChart size={20} />
            </div>
            <h1 className="auth-title">Create your account</h1>
            <p className="auth-subtitle">Get started with real-time polls in under 30 seconds</p>
          </div>

          {error && (
            <div className="auth-error-banner" role="alert">
              <span className="error-dot" />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="auth-form" noValidate>
            <div className="form-group">
              <label htmlFor="name" className="input-label">Full Name</label>
              <div className="input-wrapper">
                <input 
                  type="text" 
                  id="name" 
                  value={name} 
                  onChange={(e) => setName(e.target.value)}
                  required 
                  autoComplete="name"
                  placeholder="Alex Morgan"
                  disabled={loading}
                />
              </div>
            </div>

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
                  placeholder="alex@company.com"
                  disabled={loading}
                />
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="password" className="input-label">Password</label>
              <div className="input-wrapper password-wrapper">
                <input 
                  type={showPassword ? 'text' : 'password'} 
                  id="password" 
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)}
                  required 
                  autoComplete="new-password"
                  placeholder="At least 8 characters"
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

              {password.length > 0 && (
                <div className="strength-meter-box">
                  <div className="strength-bar-track">
                    <div className={`strength-segment ${strength >= 25 ? strengthClass : ''}`} />
                    <div className={`strength-segment ${strength >= 50 ? strengthClass : ''}`} />
                    <div className={`strength-segment ${strength >= 75 ? strengthClass : ''}`} />
                    <div className={`strength-segment ${strength === 100 ? strengthClass : ''}`} />
                  </div>
                  <div className="strength-meta">
                    <span className="strength-label-text">Security: <strong className={strengthClass}>{strengthLabel}</strong></span>
                  </div>
                </div>
              )}
            </div>

            <button 
              type="submit" 
              className="btn btn-primary auth-submit-btn" 
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="button-spinner"></span>
                  <span>Creating account...</span>
                </>
              ) : (
                <span>Create Account</span>
              )}
            </button>
          </form>

          <div className="auth-footer">
            <span>Already have an account?</span>{' '}
            <Link to="/login" className="auth-link">Sign in</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
