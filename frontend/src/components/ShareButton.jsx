import React, { useState } from 'react';
import { IconShare, IconCopy, IconCheck } from './Icons';
import Modal from './Modal';
import './ShareButton.css';

export default function ShareButton({ url, code, question }) {
  const [modalOpen, setModalOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const shareUrl = url || (typeof window !== 'undefined' ? window.location.href : '');

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2200);
    } catch (err) {
      console.error('Failed to copy', err);
    }
  };

  const handleShareClick = () => {
    setModalOpen(true);
  };

  const encodedUrl = encodeURIComponent(shareUrl);
  const encodedTitle = encodeURIComponent(question ? `Vote on: "${question}" on LivePoll` : 'Vote on this live poll!');

  return (
    <>
      <button 
        type="button"
        onClick={handleShareClick}
        className="btn btn-secondary share-trigger-btn"
        aria-label="Share this poll"
      >
        <IconShare size={15} />
        <span>Share Poll</span>
      </button>

      <Modal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Share Live Poll"
        maxWidth="480px"
      >
        <div className="share-modal-content">
          <p className="share-modal-desc">
            Invite your audience to vote instantly. Anyone with this link can cast their vote in real-time.
          </p>

          <div className="share-url-group">
            <input 
              type="text" 
              readOnly 
              value={shareUrl} 
              className="share-url-input" 
              onClick={(e) => e.target.select()}
            />
            <button 
              type="button" 
              onClick={handleCopy} 
              className={`btn ${copied ? 'btn-success-copied' : 'btn-primary'} share-copy-btn`}
            >
              {copied ? (
                <>
                  <IconCheck size={16} />
                  <span>Copied!</span>
                </>
              ) : (
                <>
                  <IconCopy size={16} />
                  <span>Copy</span>
                </>
              )}
            </button>
          </div>

          <div className="share-shortcuts">
            <span className="share-shortcuts-title">Quick Share</span>
            <div className="share-shortcuts-grid">
              <a 
                href={`https://twitter.com/intent/tweet?url=${encodedUrl}&text=${encodedTitle}`}
                target="_blank" 
                rel="noopener noreferrer"
                className="shortcut-btn x-btn"
              >
                X (Twitter)
              </a>
              <a 
                href={`https://api.whatsapp.com/send?text=${encodedTitle}%20${encodedUrl}`}
                target="_blank" 
                rel="noopener noreferrer"
                className="shortcut-btn whatsapp-btn"
              >
                WhatsApp
              </a>
              <a 
                href={`mailto:?subject=${encodedTitle}&body=${encodedTitle}%0A%0A${encodedUrl}`}
                className="shortcut-btn email-btn"
              >
                Email
              </a>
            </div>
          </div>
        </div>
      </Modal>
    </>
  );
}
