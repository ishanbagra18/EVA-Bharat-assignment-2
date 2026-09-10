import React, { useState, useEffect, useCallback } from 'react';
import { Header } from './components/Header';
import { SyncBanner } from './components/SyncBanner';
import { WindowGrid } from './components/WindowGrid';
import { AddMediaForm } from './components/AddMediaForm';
import { SyncControl } from './components/SyncControl';
import { useServerTimeOffset } from './hooks/useServerTimeOffset';
import { useSyncEvents } from './hooks/useSyncEvents';
import { fetchWindows, fetchMedia } from './api/client';
import { Radio, Plus, RefreshCw, LayoutGrid } from 'lucide-react';

export function App() {
  const [windows, setWindows] = useState([]);
  const [mediaLibrary, setMediaLibrary] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Modals
  const [showAddMediaModal, setShowAddMediaModal] = useState(false);
  const [targetWindowId, setTargetWindowId] = useState(null);
  const [showSyncModal, setShowSyncModal] = useState(false);

  // 1. Clock skew measurement
  const { offset: serverTimeOffset, isSynced } = useServerTimeOffset();

  // 2. Data Loaders
  const loadData = useCallback(async () => {
    try {
      const [winData, mediaData] = await Promise.all([
        fetchWindows(),
        fetchMedia(),
      ]);
      setWindows(winData || []);
      setMediaLibrary(mediaData || []);
      setError(null);
    } catch (err) {
      console.error('Failed to load studio data:', err);
      setError('Unable to connect to backend server on http://localhost:8085.');
    } finally {
      setLoading(false);
    }
  }, []);

  // 3. Realtime SSE listener callback for playlist updates
  const handlePlaylistUpdated = useCallback((windowId) => {
    loadData();
  }, [loadData]);

  const { syncState, isConnected } = useSyncEvents(handlePlaylistUpdated);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleOpenAddMedia = (windowId = null) => {
    setTargetWindowId(windowId);
    setShowAddMediaModal(true);
  };

  return (
    <div className="studio-app">
      <Header
        isConnected={isConnected}
        isSynced={isSynced}
        syncState={syncState}
        onTriggerSyncClick={() => setShowSyncModal(true)}
        onAddMediaClick={handleOpenAddMedia}
        onRefresh={loadData}
      />

      {/* Master Action Strip */}
      <div className="master-controls">
        <div className="master-title">
          <LayoutGrid size={18} color="var(--accent-cyan)" />
          <span>Active Channels: {windows.length} Monitors</span>
        </div>

        <div className="master-btn-group">
          <button className="btn-studio btn-studio-dark" onClick={loadData} title="Refresh Feeds">
            <RefreshCw size={14} /> Refresh
          </button>

          <button className="btn-studio btn-studio-primary" onClick={() => handleOpenAddMedia(null)}>
            <Plus size={15} /> Add Media
          </button>

          <button className="btn-studio btn-studio-override" onClick={() => setShowSyncModal(true)}>
            <Radio size={15} /> Broadcast Override Sync
          </button>
        </div>
      </div>

      <SyncBanner syncState={syncState} serverTimeOffset={serverTimeOffset} />

      {error && (
        <div style={{ background: 'rgba(255, 71, 87, 0.15)', border: '1px solid #ff4757', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem', color: '#ff6b81', fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>
          {error}
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '5rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
          <p>INITIALIZING MASTER SEQUENCER PIPELINE...</p>
        </div>
      ) : (
        <WindowGrid
          windows={windows}
          syncState={syncState}
          serverTimeOffset={serverTimeOffset}
          onAddMediaClick={handleOpenAddMedia}
        />
      )}

      {showAddMediaModal && (
        <AddMediaForm
          windows={windows}
          mediaLibrary={mediaLibrary}
          selectedWindowId={targetWindowId}
          onClose={() => setShowAddMediaModal(false)}
          onSuccess={loadData}
        />
      )}

      {showSyncModal && (
        <SyncControl
          mediaLibrary={mediaLibrary}
          onClose={() => setShowSyncModal(false)}
          onSuccess={loadData}
        />
      )}
    </div>
  );
}

export default App;
