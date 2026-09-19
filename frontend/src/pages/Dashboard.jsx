import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api';
import PollCard from '../components/PollCard';
import LoadingSpinner from '../components/LoadingSpinner';
import { IconPlus, IconChart, IconPulse, IconLock } from '../components/Icons';
import './Dashboard.css';

export default function Dashboard() {
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('all'); // 'all' | 'active' | 'closed'

  useEffect(() => {
    fetchPolls();
  }, []);

  const fetchPolls = async () => {
    try {
      const data = await api.getMyPolls();
      setPolls(data || []);
    } catch (err) {
      setError('Failed to load your polls. Please refresh.');
    } finally {
      setLoading(false);
    }
  };

  const isPollActive = (poll) => {
    const isExpired = poll.closesAt && new Date(poll.closesAt) < new Date();
    return !poll.isClosed && !isExpired;
  };

  const activePolls = polls.filter(p => isPollActive(p));
  const closedPolls = polls.filter(p => !isPollActive(p));

  const filteredPolls = filter === 'active' 
    ? activePolls 
    : filter === 'closed' 
      ? closedPolls 
      : polls;

  if (loading) return <LoadingSpinner message="Loading your dashboard..." />;

  return (
    <div className="dashboard-page container">
      {/* Page Header */}
      <div className="dashboard-header-block">
        <div className="dashboard-title-group">
          <h1 className="dashboard-main-title">Polls Dashboard</h1>
          <p className="dashboard-subtitle">
            Manage your questions, share live links, and analyze real-time voter engagement.
          </p>
        </div>
        <Link to="/create" className="btn btn-primary create-poll-cta">
          <IconPlus size={16} />
          <span>Create New Poll</span>
        </Link>
      </div>

      {/* Stats Summary Cards (Calculated from existing data) */}
      <div className="dashboard-metrics-grid">
        <div className="metric-card glass-panel">
          <div className="metric-icon-box metric-total">
            <IconChart size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-number">{polls.length}</span>
            <span className="metric-label">Total Polls Created</span>
          </div>
        </div>

        <div className="metric-card glass-panel">
          <div className="metric-icon-box metric-active">
            <IconPulse size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-number">{activePolls.length}</span>
            <span className="metric-label">Active / Live Now</span>
          </div>
        </div>

        <div className="metric-card glass-panel">
          <div className="metric-icon-box metric-closed">
            <IconLock size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-number">{closedPolls.length}</span>
            <span className="metric-label">Archived / Closed</span>
          </div>
        </div>
      </div>

      {error && (
        <div className="dashboard-error-banner" role="alert">
          <span>{error}</span>
        </div>
      )}

      {/* Tabs Filter */}
      {polls.length > 0 && (
        <div className="dashboard-filters-row">
          <div className="filter-pills" role="tablist">
            <button 
              className={`filter-pill ${filter === 'all' ? 'active' : ''}`}
              onClick={() => setFilter('all')}
              role="tab"
              aria-selected={filter === 'all'}
            >
              All Polls <span className="pill-count">{polls.length}</span>
            </button>
            <button 
              className={`filter-pill ${filter === 'active' ? 'active' : ''}`}
              onClick={() => setFilter('active')}
              role="tab"
              aria-selected={filter === 'active'}
            >
              Active <span className="pill-count">{activePolls.length}</span>
            </button>
            <button 
              className={`filter-pill ${filter === 'closed' ? 'active' : ''}`}
              onClick={() => setFilter('closed')}
              role="tab"
              aria-selected={filter === 'closed'}
            >
              Closed <span className="pill-count">{closedPolls.length}</span>
            </button>
          </div>
        </div>
      )}

      {/* Polls Content */}
      {polls.length === 0 && !error ? (
        <div className="dashboard-empty-card glass-panel-elevated animate-fade-in">
          <div className="empty-icon-wrapper">
            <IconChart size={36} />
          </div>
          <h2 className="empty-title">No polls created yet</h2>
          <p className="empty-desc">
            Launch your first live poll and start collecting instant audience responses in real-time.
          </p>
          <Link to="/create" className="btn btn-primary empty-action-btn">
            <IconPlus size={16} />
            <span>Create Your First Poll</span>
          </Link>
        </div>
      ) : filteredPolls.length === 0 ? (
        <div className="dashboard-empty-filter glass-panel">
          <p>No {filter} polls found matching your filter.</p>
        </div>
      ) : (
        <div className="polls-grid animate-fade-in">
          {filteredPolls.map(poll => (
            <PollCard key={poll.id || poll.code} poll={poll} />
          ))}
        </div>
      )}
    </div>
  );
}
