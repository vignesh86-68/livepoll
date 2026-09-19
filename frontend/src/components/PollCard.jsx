import React from 'react';
import { Link } from 'react-router-dom';
import { IconArrowRight, IconLock, IconClock } from './Icons';
import './PollCard.css';

export default function PollCard({ poll }) {
  const isExpired = poll.closesAt && new Date(poll.closesAt) < new Date();
  const isOpen = !poll.isClosed && !isExpired;

  const formattedDate = new Date(poll.createdAt).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  });

  return (
    <Link to={`/poll/${poll.code}`} className="poll-card glass-panel">
      <div className="poll-card-header">
        <span className={`poll-status-tag ${isOpen ? 'status-open' : 'status-closed'}`}>
          {isOpen ? (
            <>
              <span className="mini-pulse-dot" />
              <span>Active</span>
            </>
          ) : (
            <>
              <IconLock size={12} />
              <span>Closed</span>
            </>
          )}
        </span>
        <span className="poll-date">
          <IconClock size={13} className="date-icon" />
          <span>{formattedDate}</span>
        </span>
      </div>

      <h3 className="poll-card-question">{poll.question}</h3>

      <div className="poll-card-footer">
        <div className="poll-badges-row">
          <span className="choice-badge">
            {poll.multi ? 'Multiple Choice' : 'Single Choice'}
          </span>
          <span className="poll-code-pill">
            #{poll.code}
          </span>
        </div>

        <span className="poll-card-arrow" aria-hidden="true">
          <IconArrowRight size={15} />
        </span>
      </div>
    </Link>
  );
}
