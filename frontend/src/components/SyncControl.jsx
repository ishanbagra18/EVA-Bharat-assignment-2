import React, { useState } from 'react';
import { Radio, X } from 'lucide-react';
import { triggerSync } from '../api/client';

export function SyncControl({ mediaLibrary, onClose, onSuccess }) {
  const [selectedMediaId, setSelectedMediaId] = useState(mediaLibrary[0]?.id || 'm2');
  const selectedItem = mediaLibrary.find((m) => m.id === selectedMediaId);
  const [duration, setDuration] = useState(selectedItem?.duration_seconds || 15);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleMediaChange = (id) => {
    setSelectedMediaId(id);
    const item = mediaLibrary.find((m) => m.id === id);
    if (item) {
      setDuration(item.duration_seconds);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (!selectedMediaId) throw new Error('Please select a media item for broadcast sync');
      await triggerSync(selectedMediaId, parseInt(duration, 10));
      if (onSuccess) onSuccess();
      if (onClose) onClose();
    } catch (err) {
      setError(err.message || 'Failed to trigger broadcast sync');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="studio-modal-backdrop" onClick={onClose}>
      <div className="studio-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header-studio">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Radio size={18} style={{ color: '#ff4757' }} />
            <h3>EMERGENCY BROADCAST OVERRIDE</h3>
          </div>
          <button className="btn-studio btn-studio-dark" style={{ padding: '0.25rem 0.5rem' }} onClick={onClose}>
            <X size={16} />
          </button>
        </div>

        <p style={{ fontSize: '0.825rem', color: 'var(--text-secondary)', marginBottom: '1.25rem', lineHeight: 1.5 }}>
          Triggers an immediate broadcast override across <strong>ALL active channels</strong>. When duration expires, channels automatically resume their virtual clock schedule.
        </p>

        {error && (
          <div style={{ background: 'rgba(255, 71, 87, 0.15)', border: '1px solid #ff4757', padding: '0.65rem', borderRadius: '4px', marginBottom: '1rem', fontSize: '0.8rem', color: '#ff6b81' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="studio-form-group">
            <label>Select Override Item</label>
            <select
              className="studio-input"
              value={selectedMediaId}
              onChange={(e) => handleMediaChange(e.target.value)}
            >
              {mediaLibrary.map((m) => (
                <option key={m.id} value={m.id}>
                  [{m.id}] {m.type.toUpperCase()} ({m.duration_seconds}s)
                </option>
              ))}
            </select>
          </div>

          <div className="studio-form-group">
            <label>Override Duration (Seconds)</label>
            <input
              type="number"
              min="1"
              max="300"
              className="studio-input"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              required
            />
          </div>

          <div className="modal-footer-studio">
            <button type="button" className="btn-studio btn-studio-dark" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-studio btn-studio-override" disabled={loading}>
              <Radio size={15} /> {loading ? 'Broadcasting...' : 'Broadcast Override Now'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
