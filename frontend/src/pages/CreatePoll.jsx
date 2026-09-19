import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';
import { IconPlus, IconTrash, IconEye, IconPulse, IconCheck, IconClock, IconSparkles } from '../components/Icons';
import './CreatePoll.css';

export default function CreatePoll() {
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState([{ text: '' }, { text: '' }]);
  const [multi, setMulti] = useState(false);
  const [closesAt, setClosesAt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleAddOption = () => {
    if (options.length < 10) {
      setOptions([...options, { text: '' }]);
    }
  };

  const handleRemoveOption = (index) => {
    if (options.length > 2) {
      const newOptions = [...options];
      newOptions.splice(index, 1);
      setOptions(newOptions);
    }
  };

  const handleOptionChange = (index, value) => {
    const newOptions = [...options];
    newOptions[index].text = value;
    setOptions(newOptions);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (question.trim().length < 3 || question.trim().length > 200) {
      return setError('Question must be between 3 and 200 characters');
    }
    
    const validOptions = options.map(o => o.text.trim()).filter(Boolean);
    if (validOptions.length < 2) {
      return setError('At least two non-empty options are required');
    }

    if (new Set(validOptions).size !== validOptions.length) {
      return setError('All options must be unique');
    }

    setError('');
    setLoading(true);

    try {
      const payload = {
        question: question.trim(),
        options: validOptions,
        multi,
        ...(closesAt && { closesAt: new Date(closesAt).toISOString() })
      };
      
      const poll = await api.createPoll(payload);
      navigate(`/poll/${poll.code}`);
    } catch (err) {
      setError(err.error || 'Failed to create poll. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-poll-page container">
      <div className="create-poll-layout">
        {/* Left Column: Form Builder */}
        <div className="create-form-column">
          <div className="create-header">
            <h1 className="create-main-title">Create a Live Poll</h1>
            <p className="create-subtitle">
              Configure your question and options. Your live poll URL will be ready immediately.
            </p>
          </div>

          {error && (
            <div className="create-error-banner" role="alert">
              <span className="error-dot" />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="create-poll-form" noValidate>
            {/* Question Input */}
            <div className="form-card glass-panel">
              <div className="form-group">
                <div className="label-count-row">
                  <label htmlFor="question" className="form-section-label">Poll Question</label>
                  <span className={`char-counter ${question.length > 180 ? 'char-counter-warn' : ''}`}>
                    {question.length}/200
                  </span>
                </div>
                <textarea 
                  id="question" 
                  value={question} 
                  onChange={(e) => setQuestion(e.target.value.slice(0, 200))}
                  placeholder="e.g. Which feature should we prioritize for Q4?"
                  rows="3"
                  required
                  className="question-textarea"
                />
              </div>
            </div>

            {/* Options List */}
            <div className="form-card glass-panel">
              <div className="options-header-row">
                <label className="form-section-label">Answer Options</label>
                <span className="options-count-tag">{options.length}/10 options</span>
              </div>

              <div className="options-list-container">
                {options.map((option, index) => (
                  <div key={index} className="option-input-item">
                    <span className="option-index-badge">
                      {String(index + 1).padStart(2, '0')}
                    </span>
                    <input 
                      type="text" 
                      value={option.text} 
                      onChange={(e) => handleOptionChange(index, e.target.value)}
                      placeholder={`Option ${index + 1}`}
                      required
                      className="option-text-input"
                    />
                    {options.length > 2 && (
                      <button 
                        type="button" 
                        onClick={() => handleRemoveOption(index)}
                        className="option-remove-btn"
                        title="Remove option"
                        aria-label={`Remove option ${index + 1}`}
                      >
                        <IconTrash size={15} />
                      </button>
                    )}
                  </div>
                ))}
              </div>

              {options.length < 10 && (
                <button 
                  type="button" 
                  onClick={handleAddOption}
                  className="btn btn-secondary add-option-cta"
                >
                  <IconPlus size={15} />
                  <span>Add Another Option</span>
                </button>
              )}
            </div>

            {/* Poll Configuration / Settings */}
            <div className="form-card glass-panel">
              <h2 className="form-section-label settings-title">Poll Settings</h2>
              
              <div className="setting-toggle-row">
                <div className="setting-info">
                  <span className="setting-name">Multiple Answers</span>
                  <p className="setting-desc">Allow voters to select more than one option</p>
                </div>
                <label className="custom-switch">
                  <input 
                    type="checkbox" 
                    checked={multi} 
                    onChange={(e) => setMulti(e.target.checked)}
                  />
                  <span className="switch-slider"></span>
                </label>
              </div>

              <div className="setting-input-block">
                <label htmlFor="closesAt" className="setting-name">
                  Auto-close Date & Time <span className="optional-tag">(Optional)</span>
                </label>
                <p className="setting-desc">Automatically prevent new votes after this timestamp</p>
                <div className="datetime-input-wrapper">
                  <input 
                    type="datetime-local" 
                    id="closesAt" 
                    value={closesAt} 
                    onChange={(e) => setClosesAt(e.target.value)}
                    className="datetime-input"
                  />
                </div>
              </div>
            </div>

            <button 
              type="submit" 
              className="btn btn-primary submit-create-btn" 
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className="button-spinner"></span>
                  <span>Publishing Live Poll...</span>
                </>
              ) : (
                <>
                  <IconSparkles size={17} />
                  <span>Publish & Open Live Poll</span>
                </>
              )}
            </button>
          </form>
        </div>

        {/* Right Column: Live Interactive Card Preview */}
        <aside className="create-preview-column">
          <div className="preview-sticky-wrap">
            <div className="preview-card-header">
              <span className="preview-tag">LIVE PREVIEW</span>
              <span className="preview-type-tag">{multi ? 'Multiple Choice' : 'Single Choice'}</span>
            </div>

            <div className="preview-render-box glass-panel-elevated">
              <div className="mock-live-badge">
                <span className="mini-pulse-dot" />
                <span>LIVE • 1 watching</span>
              </div>

              <h3 className="mock-poll-title">
                {question.trim() || 'Your question will appear here...'}
              </h3>

              <div className="mock-options-list">
                {options.map((opt, i) => (
                  <div key={i} className="mock-option-item">
                    <span className={`mock-selector ${multi ? 'mock-checkbox' : 'mock-radio'}`} />
                    <span className="mock-option-text">
                      {opt.text.trim() || `Option ${i + 1}`}
                    </span>
                  </div>
                ))}
              </div>

              <div className="mock-footer">
                <span className="mock-note">Voters can submit responses anonymously</span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
