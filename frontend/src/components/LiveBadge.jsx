import React from 'react';
import { IconEye, IconLock } from './Icons';
import './LiveBadge.css';

export default function LiveBadge({ viewers = 0, status = 'open' }) {
  const isClosed = status === 'closed';

  if (isClosed) {
    return (
      <div className="live-status-pill closed" role="status" aria-label="Poll Closed">
        <IconLock size={13} className="status-icon" />
        <span className="status-text">Poll Closed</span>
      </div>
    );
  }

  return (
    <div className="live-status-pill live" role="status" aria-label={`Live Poll with ${viewers} active viewers`}>
      <span className="live-pulse-container" aria-hidden="true">
        <span className="live-pulse-ring"></span>
        <span className="live-pulse-dot"></span>
      </span>
      <span className="live-tag">LIVE</span>
      <span className="live-divider" aria-hidden="true">•</span>
      <span className="live-viewers">
        <IconEye size={13} className="eye-icon" aria-hidden="true" />
        <span className="viewer-count">{viewers}</span>
        <span className="viewer-label">{viewers === 1 ? 'watching' : 'watching'}</span>
      </span>
    </div>
  );
}
