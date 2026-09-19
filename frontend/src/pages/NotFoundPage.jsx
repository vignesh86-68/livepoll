import React from 'react';
import { Link } from 'react-router-dom';
import { IconArrowRight, IconChart } from '../components/Icons';
import './NotFoundPage.css';

export default function NotFoundPage() {
  return (
    <div className="not-found-page container">
      <div className="not-found-glow" aria-hidden="true" />
      <div className="not-found-card glass-panel-elevated animate-fade-in">
        <div className="not-found-badge">
          <IconChart size={24} />
        </div>
        <span className="not-found-code gradient-text">404</span>
        <h1 className="not-found-title">Poll or Page Not Found</h1>
        <p className="not-found-desc">
          The poll code you entered might be incorrect, expired, or the poll was deleted by its creator.
        </p>
        <div className="not-found-actions">
          <Link to="/" className="btn btn-primary">
            <span>Back to Home</span>
            <IconArrowRight size={16} />
          </Link>
          <Link to="/create" className="btn btn-secondary">
            <span>Create a New Poll</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
