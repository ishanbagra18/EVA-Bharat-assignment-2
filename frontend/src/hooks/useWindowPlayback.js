import { useState, useEffect } from 'react';

function mod(n, m) {
  return ((n % m) + m) % m;
}

export function useWindowPlayback(window, syncState, serverTimeOffset = 0) {
  const [current, setCurrent] = useState(null);

  useEffect(() => {
    if (!window) return;

    const tick = () => {
      const now = Date.now() + serverTimeOffset;

      // 1. Evaluate Sync Override First
      if (syncState?.active && now < syncState.endsAt) {
        const offsetSec = Math.max(0, (now - syncState.startedAt) / 1000);
        const remainingSec = Math.max(0, (syncState.endsAt - now) / 1000);
        setCurrent({
          media: syncState.media,
          offsetSeconds: offsetSec,
          isSync: true,
          syncRemainingSec: remainingSec,
          activeEntryId: null,
        });
        return;
      }

      // 2. Fall back cleanly to Deterministic Virtual Clock calculation
      const playlist = window.playlist || [];
      const playlistTotal = playlist.reduce(
        (sum, entry) => sum + (entry.media?.duration_seconds || 0),
        0
      );

      if (playlistTotal === 0) {
        setCurrent({ media: null, isSync: false, offsetSeconds: 0, activeEntryId: null });
        return;
      }

      const cycleEpochMs = window.cycle_epoch
        ? new Date(window.cycle_epoch).getTime()
        : new Date('2026-01-01T00:00:00Z').getTime();

      const cycleLenSec = window.cycle_length_seconds || 18000;
      const elapsedInCycleSec = mod((now - cycleEpochMs) / 1000, cycleLenSec);
      const positionInLoopSec = mod(elapsedInCycleSec, playlistTotal);

      let cursor = 0;
      for (const entry of playlist) {
        const d = entry.media?.duration_seconds || 0;
        if (d <= 0) continue;

        if (positionInLoopSec < cursor + d) {
          setCurrent({
            media: entry.media,
            offsetSeconds: positionInLoopSec - cursor,
            isSync: false,
            activeEntryId: entry.id,
            activeMediaId: entry.media_id,
          });
          return;
        }
        cursor += d;
      }

      // Floating boundary fallback to last item
      if (playlist.length > 0) {
        const last = playlist[playlist.length - 1];
        setCurrent({
          media: last.Media || last.media,
          offsetSeconds: last.media?.duration_seconds || 0,
          isSync: false,
          activeEntryId: last.id,
        });
      }
    };

    tick();
    const intervalId = setInterval(tick, 500); // 500ms re-calculation tick
    return () => clearInterval(intervalId);
  }, [window, syncState, serverTimeOffset]);

  return current;
}
