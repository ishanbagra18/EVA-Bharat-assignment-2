import React, { useState, useEffect } from 'react';
import { Radio, AlertTriangle } from 'lucide-react';

export function SyncBanner({ syncState, serverTimeOffset }) {
  const [remainingSec, setRemainingSec] = useState(0);

  useEffect(() => {
    if (!syncState?.active || !syncState.endsAt) {
      setRemainingSec(0);
      return;
    }

    const updateTimer = () => {
      const now = Date.now() + serverTimeOffset;
      const left = Math.max(0, (syncState.endsAt - now) / 1000);
      setRemainingSec(left);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 100);
    return () => clearInterval(interval);
  }, [syncState, serverTimeOffset]);

  if (!syncState?.active || remainingSec <= 0) return null;

  return (
    <div className="broadcast-override-alert">
      <div className="override-info">
        <Radio size={20} className="animate-pulse" />
        <span>MASTER BROADCAST OVERRIDE ACTIVE — Item #{syncState.media_id || syncState.media?.id} ({syncState.media?.type})</span>
      </div>

      <div className="override-timer">
        RESUMING SCHEDULE IN {remainingSec.toFixed(1)}s
      </div>
    </div>
  );
}
