import React, { useState } from 'react';
import { X, Plus } from 'lucide-react';
import { addMediaToWindow } from '../api/client';

export function AddMediaForm({ windows, mediaLibrary, selectedWindowId, onClose, onSuccess }) {
  const [windowId, setWindowId] = useState(selectedWindowId || (windows[0]?.id || ''));
  const [mode, setMode] = useState('library');
  const [selectedMediaId, setSelectedMediaId] = useState(mediaLibrary[0]?.id || '');
  
  const [type, setType] = useState('image');
  const [url, setUrl] = useState('https://images.unsplash.com/photo-1550684848-fac1c5b4e853?w=1200');
  const [duration, setDuration] = useState(15);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      let payload = {};
      if (mode === 'library') {
        if (!selectedMediaId) throw new Error('Please select a media item from library');
        payload = { media_id: selectedMediaId };
      } else {
        if (type !== 'blank' && !url) throw new Error('URL is required for image/video items');
        if (!duration || duration <= 0) throw new Error('Duration must be greater than 0');
        payload = {
          type,
          url: type === 'blank' ? null : url,
          duration_seconds: parseInt(duration, 10),
        };
      }

      await addMediaToWindow(windowId, payload);
      onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to add media item');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="studio-modal-backdrop" onClick={onClose}>
      <div className="studio-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header-studio">
          <h3>ADD MEDIA ITEM TO SEQUENCE</h3>
          <button className="btn-studio btn-studio-dark" style={{ padding: '0.25rem 0.5rem' }} onClick={onClose}>
            <X size={16} />
          </button>
        </div>

        {error && (
          <div style={{ background: 'rgba(255, 71, 87, 0.15)', border: '1px solid #ff4757', padding: '0.65rem', borderRadius: '4px', marginBottom: '1rem', fontSize: '0.8rem', color: '#ff6b81' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="studio-form-group">
            <label>Target Channel Monitor</label>
            <select
              className="studio-input"
              value={windowId}
              onChange={(e) => setWindowId(e.target.value)}
            >
              {windows.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name.toUpperCase()} [{w.id}]
                </option>
              ))}
            </select>
          </div>

          <div className="studio-form-group">
            <label>Media Source Mode</label>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                type="button"
                className={`btn-studio ${mode === 'library' ? 'btn-studio-primary' : 'btn-studio-dark'}`}
                style={{ flex: 1, justifyContent: 'center' }}
                onClick={() => setMode('library')}
              >
                Shared Library
              </button>
              <button
                type="button"
                className={`btn-studio ${mode === 'custom' ? 'btn-studio-primary' : 'btn-studio-dark'}`}
                style={{ flex: 1, justifyContent: 'center' }}
                onClick={() => setMode('custom')}
              >
                New Item
              </button>
            </div>
          </div>

          {mode === 'library' ? (
            <div className="studio-form-group">
              <label>Select Library Item</label>
              <select
                className="studio-input"
                value={selectedMediaId}
                onChange={(e) => setSelectedMediaId(e.target.value)}
              >
                {mediaLibrary.map((m) => (
                  <option key={m.id} value={m.id}>
                    [{m.id}] {m.type.toUpperCase()} ({m.duration_seconds}s)
                  </option>
                ))}
              </select>
            </div>
          ) : (
            <>
              <div className="studio-form-group">
                <label>Media Type</label>
                <select
                  className="studio-input"
                  value={type}
                  onChange={(e) => setType(e.target.value)}
                >
                  <option value="image">Image Clip</option>
                  <option value="video">Video Clip</option>
                  <option value="blank">Blank Slot</option>
                </select>
              </div>

              {type !== 'blank' && (
                <div className="studio-form-group">
                  <label>Media Source URL</label>
                  <input
                    type="url"
                    className="studio-input"
                    placeholder="https://..."
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    required
                  />
                </div>
              )}

              <div className="studio-form-group">
                <label>Display Duration (Seconds)</label>
                <input
                  type="number"
                  min="1"
                  className="studio-input"
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                  required
                />
              </div>
            </>
          )}

          <div className="modal-footer-studio">
            <button type="button" className="btn-studio btn-studio-dark" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-studio btn-studio-primary" disabled={loading}>
              <Plus size={15} /> {loading ? 'Adding...' : 'Add to Sequence'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
