import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { 
  IconPulse, 
  IconZap, 
  IconShare, 
  IconShield, 
  IconArrowRight, 
  IconCheck, 
  IconChart, 
  IconEye, 
  IconClock 
} from '../components/Icons';
import './LandingPage.css';

export default function LandingPage() {
  // Interactive product preview state
  const [demoSelected, setDemoSelected] = useState(1);
  const [demoVotes, setDemoVotes] = useState({ 0: 42, 1: 185, 2: 73 });

  const demoOptions = [
    { text: 'Go (Gin + Goroutines)' },
    { text: 'Node.js (TypeScript)' },
    { text: 'Python (FastAPI)' }
  ];

  const handleDemoVote = (idx) => {
    if (demoSelected === idx) return;
    setDemoVotes(prev => {
      const updated = { ...prev };
      if (demoSelected !== null) updated[demoSelected] = Math.max(0, updated[demoSelected] - 1);
      updated[idx] = (updated[idx] || 0) + 1;
      return updated;
    });
    setDemoSelected(idx);
  };

  const totalDemoVotes = Object.values(demoVotes).reduce((a, b) => a + b, 0);

  return (
    <div className="landing-page">
      {/* Background Decorative Mesh */}
      <div className="mesh-glow mesh-glow-top" aria-hidden="true" />
      <div className="mesh-glow mesh-glow-bottom" aria-hidden="true" />

      {/* Hero Section */}
      <section className="hero-section container">
        <div className="hero-announcement-badge">
          <span className="announcement-dot" />
          <span className="announcement-text">Powered by Redis Cloud & Go WebSockets</span>
        </div>

        <h1 className="hero-headline">
          Create live polls.<br />
          <span className="gradient-text">Get instant responses.</span>
        </h1>

        <p className="hero-subheadline">
          Capture live audience decisions, gather conference feedback, and broadcast real-time poll tallies with sub-millisecond Redis Pub/Sub synchronization.
        </p>

        <div className="hero-cta-group">
          <Link to="/create" className="btn btn-primary hero-btn-main">
            <span>Create a Poll</span>
            <IconArrowRight size={17} />
          </Link>
          <a href="#how-it-works" className="btn btn-secondary hero-btn-secondary">
            <span>See How it Works</span>
          </a>
        </div>

        {/* Live Interactive Product Preview Card */}
        <div className="product-preview-container">
          <div className="product-preview-card glass-panel-elevated">
            <div className="preview-header">
              <div className="preview-badge-live">
                <span className="preview-live-dot"></span>
                <span className="preview-live-label">LIVE DEMO</span>
                <span className="preview-divider">•</span>
                <span className="preview-viewers">
                  <IconEye size={13} />
                  <span>248 watching</span>
                </span>
              </div>
              <span className="preview-code-tag">#LIVE-PREVIEW</span>
            </div>

            <div className="preview-body">
              <h2 className="preview-question">Which backend stack do you prefer for real-time systems?</h2>
              <p className="preview-instruction">Click an option below to test live interactive tallying:</p>

              <div className="preview-options">
                {demoOptions.map((opt, idx) => {
                  const count = demoVotes[idx] || 0;
                  const pct = totalDemoVotes > 0 ? Math.round((count / totalDemoVotes) * 100) : 0;
                  const isSelected = demoSelected === idx;

                  return (
                    <div 
                      key={idx}
                      className={`preview-option-card ${isSelected ? 'selected' : ''}`}
                      onClick={() => handleDemoVote(idx)}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => e.key === 'Enter' && handleDemoVote(idx)}
                    >
                      <div className="preview-option-info">
                        <div className="preview-option-left">
                          <span className={`preview-radio ${isSelected ? 'active' : ''}`}>
                            {isSelected && <IconCheck size={12} />}
                          </span>
                          <span className="preview-option-name">{opt.text}</span>
                        </div>
                        <div className="preview-option-stats">
                          <span className="preview-vote-count">{count} votes</span>
                          <span className="preview-percentage">{pct}%</span>
                        </div>
                      </div>
                      <div className="preview-bar-track">
                        <div className="preview-bar-fill" style={{ width: `${pct}%` }} />
                      </div>
                    </div>
                  );
                })}
              </div>

              <div className="preview-footer">
                <span className="preview-total">{totalDemoVotes} total votes cast</span>
                <span className="preview-live-indicator">
                  <IconPulse size={14} className="pulse-icon" />
                  <span>Sub-millisecond socket sync</span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Feature Pillars Section */}
      <section id="features" className="features-section container">
        <div className="section-header">
          <span className="section-pill">ENGINEERED FOR SCALE</span>
          <h2 className="section-title">Built for high-concurrency live audiences</h2>
          <p className="section-subtitle">
            Engineered with high performance technologies for reliable, low-latency live interaction.
          </p>
        </div>

        <div className="features-grid">
          <div className="feature-card glass-panel">
            <div className="feature-icon-wrapper">
              <IconZap size={22} className="feature-icon" />
            </div>
            <h3 className="feature-title">Sub-Millisecond Redis Engine</h3>
            <p className="feature-desc">
              Every vote triggers an atomic increment in Redis memory, fanning out instant updates across distributed pub/sub channels without database bottleneck.
            </p>
          </div>

          <div className="feature-card glass-panel">
            <div className="feature-icon-wrapper">
              <IconPulse size={22} className="feature-icon" />
            </div>
            <h3 className="feature-title">Bi-Directional WebSockets</h3>
            <p className="feature-desc">
              Persistent WebSocket channels keep all connected viewers in sync. Viewers see tallies glide automatically without ever refreshing their browser.
            </p>
          </div>

          <div className="feature-card glass-panel">
            <div className="feature-icon-wrapper">
              <IconShare size={22} className="feature-icon" />
            </div>
            <h3 className="feature-title">Zero-Friction Voter Access</h3>
            <p className="feature-desc">
              Voters simply click a clean 6-character link. No account creation, app downloads, or passwords required to submit real-time votes.
            </p>
          </div>

          <div className="feature-card glass-panel">
            <div className="feature-icon-wrapper">
              <IconShield size={22} className="feature-icon" />
            </div>
            <h3 className="feature-title">Anti-Tamper Vote Tracking</h3>
            <p className="feature-desc">
              Integrated voter fingerprinting and stateful validation prevent double voting while supporting both single-answer and multi-select formats.
            </p>
          </div>
        </div>
      </section>

      {/* How it Works Section */}
      <section id="how-it-works" className="steps-section container">
        <div className="section-header">
          <span className="section-pill">HOW IT WORKS</span>
          <h2 className="section-title">From idea to live feedback in 3 steps</h2>
        </div>

        <div className="steps-grid">
          <div className="step-card glass-panel">
            <span className="step-number">01</span>
            <h3 className="step-title">Create your Question</h3>
            <p className="step-desc">
              Add custom options, configure single or multiple choices, and set an optional automated closing timer.
            </p>
          </div>

          <div className="step-card glass-panel">
            <span className="step-number">02</span>
            <h3 className="step-title">Share the Live Link</h3>
            <p className="step-desc">
              Distribute your poll URL or code across chat, slides, or social media. Anyone can join instantly from any device.
            </p>
          </div>

          <div className="step-card glass-panel">
            <span className="step-number">03</span>
            <h3 className="step-title">Watch Votes Stream In</h3>
            <p className="step-desc">
              Present the results live on stage or screen. Watch animated bars shift dynamically as thousands of votes arrive.
            </p>
          </div>
        </div>
      </section>

      {/* Bottom CTA Banner */}
      <section className="cta-banner-section container">
        <div className="cta-banner glass-panel-elevated">
          <div className="cta-content">
            <h2 className="cta-title">Ready to engage your audience?</h2>
            <p className="cta-subtitle">
              Start creating interactive live polls in seconds. Free for creators and attendees.
            </p>
          </div>
          <div className="cta-actions">
            <Link to="/create" className="btn btn-primary cta-btn">
              <span>Create Your First Poll</span>
              <IconArrowRight size={17} />
            </Link>
          </div>
        </div>
      </section>

      {/* Polished Footer */}
      <footer className="landing-footer">
        <div className="container footer-content">
          <div className="footer-brand">
            <div className="brand-logo-glow">
              <IconChart size={18} />
            </div>
            <span className="footer-brand-name">
              Live<span className="gradient-text">Poll</span>
            </span>
          </div>

          <div className="footer-status">
            <span className="footer-status-dot" />
            <span className="footer-status-text">All Systems Operational</span>
          </div>

          <p className="footer-copy">
            © {new Date().getFullYear()} LivePoll. Full Stack Real-Time Polling Platform.
          </p>
        </div>
      </footer>
    </div>
  );
}
