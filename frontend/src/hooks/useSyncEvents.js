import { useState, useEffect, useCallback } from 'react';
import { getEventsStreamUrl, fetchSyncStatus } from '../api/client';

export function useSyncEvents(onPlaylistUpdated) {
  const [syncState, setSyncState] = useState({ active: false });
  const [isConnected, setIsConnected] = useState(false);

  const refreshSyncStatus = useCallback(async () => {
    try {
      const status = await fetchSyncStatus();
      setSyncState({
        active: status.active,
        media: status.media,
        mediaId: status.media_id,
        startedAt: status.started_at ? new Date(status.started_at).getTime() : 0,
        endsAt: status.ends_at ? new Date(status.ends_at).getTime() : 0,
        durationSeconds: status.duration_seconds,
      });
    } catch (err) {
      console.warn('Failed to fetch sync status:', err);
    }
  }, []);

  useEffect(() => {
    refreshSyncStatus();

    const streamUrl = getEventsStreamUrl();
    const eventSource = new EventSource(streamUrl);

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.addEventListener('sync', (e) => {
      try {
        const data = JSON.parse(e.data);
        setSyncState({
          active: data.active,
          media: data.media,
          mediaId: data.media_id,
          startedAt: data.started_at ? new Date(data.started_at).getTime() : 0,
          endsAt: data.ends_at ? new Date(data.ends_at).getTime() : 0,
          durationSeconds: data.duration_seconds,
        });
      } catch (err) {
        console.error('Error parsing sync SSE event:', err);
      }
    });

    eventSource.addEventListener('playlist_updated', (e) => {
      if (onPlaylistUpdated) {
        try {
          const data = JSON.parse(e.data);
          onPlaylistUpdated(data.window_id);
        } catch (err) {
          onPlaylistUpdated(null);
        }
      }
    });

    eventSource.onerror = () => {
      setIsConnected(false);
      // EventSource automatically retries connections
    };

    return () => {
      eventSource.close();
    };
  }, [onPlaylistUpdated, refreshSyncStatus]);

  return { syncState, isConnected, refreshSyncStatus };
}
