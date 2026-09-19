import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api';
import { useWebSocket } from '../hooks/useWebSocket';
import { useAuth } from '../context/AuthContext';
import ResultBar from '../components/ResultBar';
import LiveBadge from '../components/LiveBadge';
import ShareButton from '../components/ShareButton';
import LoadingSpinner from '../components/LoadingSpinner';
import Modal from '../components/Modal';
import { IconCheck, IconTrash, IconLock, IconPulse, IconUser } from '../components/Icons';
import './PollPage.css';

export default function PollPage() {
  const { code } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  
  const [poll, setPoll] = useState(null);
  const [hasVoted, setHasVoted] = useState(false);
  const [selectedOptions, setSelectedOptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [voting, setVoting] = useState(false);
  const [error, setError] = useState('');

  // Confirmation Modals
  const [showCloseModal, setShowCloseModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [modalLoading, setModalLoading] = useState(false);

  // WebSocket hook (Strictly untouched)
  const { tally, viewers, status } = useWebSocket(code);

  useEffect(() => {
    fetchPoll();
  }, [code]);

  const fetchPoll = async () => {
    try {
      const data = await api.getPoll(code);
      setPoll(data.poll);
      setHasVoted(data.hasVoted);
    } catch (err) {
      if (err.status === 404) {
        navigate('/404');
      } else {
        setError('Failed to load poll. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const toggleOption = (id) => {
    if (poll.multi) {
      if (selectedOptions.includes(id)) {
        setSelectedOptions(selectedOptions.filter(optId => optId !== id));
      } else {
        setSelectedOptions([...selectedOptions, id]);
      }
    } else {
      setSelectedOptions([id]);
    }
  };

  const handleVote = async () => {
    if (selectedOptions.length === 0) return;
    
    setVoting(true);
    setError('');
    
    try {
      await api.vote(code, selectedOptions);
      setHasVoted(true);
    } catch (err) {
      setError(err.error || 'Failed to submit vote');
    } finally {
      setVoting(false);
    }
  };

  const handleConfirmClose = async () => {
    setModalLoading(true);
    try {
      await api.closePoll(code);
      setShowCloseModal(false);
    } catch (err) {
      setError(err.error || 'Failed to close poll');
    } finally {
      setModalLoading(false);
    }
  };

  const handleConfirmDelete = async () => {
    setModalLoading(true);
    try {
      await api.deletePoll(code);
      setShowDeleteModal(false);
      navigate('/dashboard');
    } catch (err) {
      setError(err.error || 'Failed to delete poll');
    } finally {
      setModalLoading(false);
    }
  };

  if (loading) return <LoadingSpinner message="Connecting to live poll session..." />;
  if (!poll) return null;

  const isOwner = user && poll.ownerId === user.id;
  const isClosed = status === 'closed' || poll.isClosed || (poll.closesAt && new Date(poll.closesAt) < new Date());
  
  // Decide whether to show voting form or results
  const showResults = hasVoted || isClosed || isOwner;

  // Calculate highest vote count to highlight leading option
  let highestVoteCount = 0;
  if (tally && tally.counts) {
    Object.values(tally.counts).forEach(count => {
      if (count > highestVoteCount) highestVoteCount = count;
    });
  }

  return (
    <div className="poll-page container">
      {/* Top Header Bar */}
      <div className="poll-top-bar">
        <LiveBadge viewers={viewers} status={isClosed ? 'closed' : 'open'} />
        <div className="poll-top-actions">
          <ShareButton 
            code={poll.code} 
            question={poll.question}
          />
        </div>
      </div>
      
      {error && (
        <div className="poll-error-banner" role="alert">
          <span>{error}</span>
        </div>
      )}

      {/* Main Poll Card */}
      <div className="poll-main-panel glass-panel-elevated animate-fade-in">
        <div className="poll-header-info">
          <div className="poll-author-tag">
            <span className="author-avatar-dot">
              <IconUser size={12} />
            </span>
            <span>Created by <strong className="author-name">{poll.ownerName}</strong></span>
          </div>

          <h1 className="poll-headline">{poll.question}</h1>

          <div className="poll-meta-strip">
            <span className="poll-type-badge">
              {poll.multi ? 'Multiple Choice (Select all that apply)' : 'Single Choice (Select one)'}
            </span>
            <span className="poll-code-display">Code: #{poll.code}</span>
          </div>
        </div>
        
        <div className="poll-card-body">
          {showResults ? (
            /* Results View */
            <div className="results-view animate-fade-in">
              <div className="results-status-header">
                <div className="results-live-indicator">
                  <IconPulse size={16} className="pulse-signal-icon" />
                  <span>Real-time tally</span>
                </div>
                <div className="total-votes-pill">
                  <strong className="total-count-number">{tally.total}</strong>{' '}
                  <span>{tally.total === 1 ? 'total response' : 'total responses'}</span>
                </div>
              </div>
              
              <div className="results-list">
                {poll.options.map(option => {
                  const count = (tally.counts && tally.counts[option.id]) || 0;
                  const isLeader = count > 0 && count === highestVoteCount;
                  return (
                    <ResultBar 
                      key={option.id}
                      option={option}
                      count={count}
                      total={tally.total}
                      isLeader={isLeader}
                    />
                  );
                })}
              </div>
              
              {isClosed && (
                <div className="closed-banner glass-panel">
                  <IconLock size={18} className="closed-banner-icon" />
                  <div className="closed-banner-content">
                    <span className="closed-banner-title">This poll is closed</span>
                    <span className="closed-banner-subtitle">Voting has ended. Above are the final recorded results.</span>
                  </div>
                </div>
              )}
            </div>
          ) : (
            /* Voting View */
            <div className="voting-view animate-fade-in">
              <div className="options-selection-grid">
                {poll.options.map(option => {
                  const isSelected = selectedOptions.includes(option.id);
                  return (
                    <button
                      key={option.id}
                      type="button"
                      className={`interactive-option-card ${isSelected ? 'selected' : ''}`}
                      onClick={() => toggleOption(option.id)}
                      role={poll.multi ? 'checkbox' : 'radio'}
                      aria-checked={isSelected}
                    >
                      <div className={`option-selector-box ${poll.multi ? 'checkbox-style' : 'radio-style'}`}>
                        {isSelected && (
                          poll.multi ? <IconCheck size={14} className="check-icon" /> : <div className="radio-dot" />
                        )}
                      </div>
                      <span className="option-label-text">{option.text}</span>
                    </button>
                  );
                })}
              </div>
              
              <button 
                type="button"
                className="btn btn-primary submit-vote-cta"
                onClick={handleVote}
                disabled={voting || selectedOptions.length === 0}
              >
                {voting ? (
                  <>
                    <span className="button-spinner"></span>
                    <span>Recording vote in Redis...</span>
                  </>
                ) : (
                  <span>Submit Your Vote</span>
                )}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Owner Management Panel */}
      {isOwner && (
        <div className="owner-management-panel glass-panel animate-fade-in">
          <div className="owner-panel-header">
            <h3 className="owner-panel-title">Host Controls</h3>
            <span className="owner-status-desc">You are the author of this poll</span>
          </div>
          <div className="owner-actions-row">
            {!isClosed && (
              <button 
                type="button" 
                className="btn btn-secondary owner-btn" 
                onClick={() => setShowCloseModal(true)}
              >
                <IconLock size={15} />
                <span>Close Voting</span>
              </button>
            )}
            <button 
              type="button" 
              className="btn btn-danger owner-btn" 
              onClick={() => setShowDeleteModal(true)}
            >
              <IconTrash size={15} />
              <span>Delete Poll</span>
            </button>
          </div>
        </div>
      )}

      {/* Confirmation Modal: Close Poll */}
      <Modal
        isOpen={showCloseModal}
        onClose={() => !modalLoading && setShowCloseModal(false)}
        title="Close this poll?"
      >
        <p>
          Are you sure you want to close <strong>"{poll.question}"</strong>? Voters will no longer be able to cast new responses, and final tallies will remain visible.
        </p>
        <div className="modal-actions">
          <button 
            type="button" 
            className="btn btn-secondary" 
            onClick={() => setShowCloseModal(false)}
            disabled={modalLoading}
          >
            Cancel
          </button>
          <button 
            type="button" 
            className="btn btn-primary" 
            onClick={handleConfirmClose}
            disabled={modalLoading}
          >
            {modalLoading ? 'Closing...' : 'Close Poll'}
          </button>
        </div>
      </Modal>

      {/* Confirmation Modal: Delete Poll */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => !modalLoading && setShowDeleteModal(false)}
        title="Delete this poll permanently?"
      >
        <p>
          Are you sure you want to delete <strong>"{poll.question}"</strong>? This will permanently remove the poll, all vote records, and real-time tallies. This action cannot be undone.
        </p>
        <div className="modal-actions">
          <button 
            type="button" 
            className="btn btn-secondary" 
            onClick={() => setShowDeleteModal(false)}
            disabled={modalLoading}
          >
            Cancel
          </button>
          <button 
            type="button" 
            className="btn btn-danger" 
            onClick={handleConfirmDelete}
            disabled={modalLoading}
          >
            {modalLoading ? 'Deleting...' : 'Permanently Delete'}
          </button>
        </div>
      </Modal>
    </div>
  );
}
