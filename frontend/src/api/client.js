const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8085';

export async function fetchHealth() {
  const res = await fetch(`${API_BASE_URL}/health`);
  if (!res.ok) throw new Error('Failed to fetch health status');
  return res.json();
}

export async function fetchWindows() {
  const res = await fetch(`${API_BASE_URL}/windows`);
  if (!res.ok) throw new Error('Failed to fetch windows');
  return res.json();
}

export async function fetchMedia() {
  const res = await fetch(`${API_BASE_URL}/media`);
  if (!res.ok) throw new Error('Failed to fetch media library');
  return res.json();
}

export async function createMedia(item) {
  const res = await fetch(`${API_BASE_URL}/media`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(item),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create media item');
  }
  return res.json();
}

export async function addMediaToWindow(windowId, payload) {
  const res = await fetch(`${API_BASE_URL}/windows/${windowId}/media`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to add media to window');
  }
  return res.json();
}

export async function triggerSync(mediaId, durationSeconds) {
  const res = await fetch(`${API_BASE_URL}/sync`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ media_id: mediaId, duration_seconds: durationSeconds }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to trigger sync');
  }
  return res.json();
}

export async function fetchSyncStatus() {
  const res = await fetch(`${API_BASE_URL}/sync/status`);
  if (!res.ok) throw new Error('Failed to fetch sync status');
  return res.json();
}

export function getEventsStreamUrl() {
  return `${API_BASE_URL}/events`;
}
