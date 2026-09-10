import React, { useState, useEffect } from 'react';
import { Radio, Plus, RefreshCw, Activity, Cpu } from 'lucide-react';

export function Header({ isConnected, isSynced, syncState, onTriggerSyncClick, onAddMediaClick, onRefresh }) {
  const [utcTime, setUtcTime] = useState('');

  useEffect(() => {
    const tick = () => {
      const d = new Date();
      setUtcTime(d.toISOString().slice(11, 19) + ' UTC');
    };
    tick();
    const interval = setInterval(tick, 1000);
    return () => clearInterval(interval);
  }, []);

  const isSyncActive = syncState?.active;

  return (
    <header className="studio-header">
      <div className="header-brand">
        <div className="tally-box">
          {isSyncActive ? 'OVERRIDE' : 'ON AIR'}
        </div>
        <div className="header-title-group">
          <h1>Master Broadcast Sequencer</h1>
          <p>Deterministic Virtual Clock Control Engine • 5-Hr Cycle</p>
        </div>
      </div>

      <div className="header-telemetry">
        <div className="telemetry-item">
          <span className="telemetry-label">Studio Timecode</span>
          <span className="telemetry-value">{utcTime}</span>
        </div>

        <div className="telemetry-item">
          <span className="telemetry-label">Sync Protocol</span>
          <span className={`telemetry-value ${isSynced ? 'green' : ''}`}>
            {isSynced ? 'LOCKED (0ms)' : 'SYNCING'}
          </span>
        </div>

        <div className="telemetry-item">
          <span className="telemetry-label">Push Stream</span>
          <span className={`telemetry-value ${isConnected ? 'green' : ''}`}>
            {isConnected ? 'LIVE (SSE)' : 'DISCONNECTED'}
          </span>
        </div>
      </div>
    </header>
  );
}
