import React from 'react';
import { IconTrending } from './Icons';
import './ResultBar.css';

export default function ResultBar({ option, count = 0, total = 0, isLeader = false }) {
  const percentage = total > 0 ? Math.round((count / total) * 100) : 0;
  
  return (
    <div className={`result-bar-wrapper ${isLeader && total > 0 ? 'is-leader' : ''}`}>
      <div className="result-bar-info">
        <div className="result-text-group">
          <span className="result-bar-text">{option.text}</span>
          {isLeader && total > 0 && count > 0 && (
            <span className="leader-badge">
              <IconTrending size={12} />
              <span>Leading</span>
            </span>
          )}
        </div>
        <div className="result-bar-stats">
          <span className="result-vote-count">{count} {count === 1 ? 'vote' : 'votes'}</span>
          <span className="result-percentage">{percentage}%</span>
        </div>
      </div>

      <div 
        className="result-bar-track"
        role="progressbar"
        aria-valuenow={percentage}
        aria-valuemin="0"
        aria-valuemax="100"
        aria-label={`${option.text}: ${percentage}% with ${count} votes`}
      >
        <div 
          className="result-bar-fill" 
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}
