import React from 'react';
import { MediaWindow } from './MediaWindow';

export function WindowGrid({ windows, syncState, serverTimeOffset, onAddMediaClick }) {
  if (!windows || windows.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: '4rem 2rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
        <h3>NO MONITOR CHANNELS CONNECTED</h3>
        <p style={{ fontSize: '0.85rem', marginTop: '0.5rem' }}>
          Check SQLite database connection or start backend server.
        </p>
      </div>
    );
  }

  return (
    <div className="monitors-grid">
      {windows.map((win) => (
        <MediaWindow
          key={win.id}
          window={win}
          syncState={syncState}
          serverTimeOffset={serverTimeOffset}
          onAddMediaClick={onAddMediaClick}
        />
      ))}
    </div>
  );
}
