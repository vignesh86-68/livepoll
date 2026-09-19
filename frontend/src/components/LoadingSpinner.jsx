import React from 'react';
import './LoadingSpinner.css';

export default function LoadingSpinner({ message = 'Loading...' }) {
  return (
    <div className="spinner-container" role="status" aria-live="polite">
      <div className="spinner-orbit">
        <div className="spinner-ring"></div>
        <div className="spinner-core"></div>
      </div>
      {message && <span className="spinner-message">{message}</span>}
    </div>
  );
}
