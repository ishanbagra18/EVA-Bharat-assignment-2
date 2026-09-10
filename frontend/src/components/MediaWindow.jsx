import React, { useRef, useEffect } from 'react';
import { Film, Image, EyeOff, Plus, Play, Radio } from 'lucide-react';
import { useWindowPlayback } from '../hooks/useWindowPlayback';

export function MediaWindow({ window, syncState, serverTimeOffset, onAddMediaClick }) {
  const playback = useWindowPlayback(window, syncState, serverTimeOffset);
  const videoRef = useRef(null);

  const media = playback?.media;
  const offsetSeconds = playback?.offsetSeconds || 0;
  const isSync = playback?.isSync || false;
  const duration = media?.duration_seconds || 1;

  useEffect(() => {
    if (media?.type === 'video' && videoRef.current) {
      const v = videoRef.current;
      const targetTime = offsetSeconds % duration;
      if (Math.abs(v.currentTime - targetTime) > 0.75) {
        v.currentTime = targetTime;
      }
    }
  }, [media, offsetSeconds, duration]);

  const renderMediaContent = () => {
    if (!media) {
      return (
        <div className="empty-signal">
          <EyeOff size={32} opacity={0.4} />
          <span>NO FEED CONFIGURED</span>
        </div>
      );
    }

    if (media.type === 'blank') {
      return (
        <div className="empty-signal">
          <EyeOff size={36} style={{ color: 'var(--text-muted)' }} />
          <span style={{ fontWeight: 700, color: '#fff' }}>BLANK SLOT ACTIVE</span>
          <span style={{ opacity: 0.6 }}>Duration: {media.duration_seconds}s</span>
        </div>
      );
    }

    if (media.type === 'image') {
      return (
        <img
          src={media.url}
          alt={media.id || 'Media'}
          onError={(e) => {
            e.target.src = 'https://images.unsplash.com/photo-1579546929518-9e396f3cc809?w=1200';
          }}
        />
      );
    }

    if (media.type === 'video') {
      return (
        <video
          ref={videoRef}
          src={media.url}
          autoPlay
          muted
          loop
          playsInline
        />
      );
    }

    return null;
  };

  const getIcon = (type) => {
    if (type === 'image') return <Image size={12} />;
    if (type === 'video') return <Film size={12} />;
    return <EyeOff size={12} />;
  };

  return (
    <div className={`monitor-frame ${isSync ? 'on-sync' : ''}`}>
      {/* Monitor Control Bar */}
      <div className="monitor-bar">
        <div className="channel-tag">
          <div className={`tally-indicator ${isSync ? 'sync-active' : ''}`} />
          <span>{window.name.toUpperCase()}</span>
          <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
            [{window.id}]
          </span>
        </div>

        <button
          className="btn-studio btn-studio-dark"
          style={{ padding: '0.3rem 0.6rem', fontSize: '0.75rem' }}
          onClick={() => onAddMediaClick(window.id)}
        >
          <Plus size={13} /> Add Item
        </button>
      </div>

      {/* Monitor Viewport */}
      <div className="monitor-viewport">
        {renderMediaContent()}

        {/* Viewport Header Tag */}
        <div className="viewport-format-tag">
          {isSync ? 'LIVE OVERRIDE' : media ? `${media.type.toUpperCase()} • ${media.duration_seconds}S` : 'NO SIGNAL'}
        </div>

        {/* Timecode Readout */}
        {media && (
          <div className={`viewport-timecode ${isSync ? 'sync-mode' : ''}`}>
            TC {offsetSeconds.toFixed(1)}s / {media.duration_seconds}.0s
          </div>
        )}
      </div>

      {/* Track Sequence Bar */}
      <div className="monitor-footer">
        <div className="sequence-info-bar">
          <span>SEQUENCE LOOP ({window.playlist?.length || 0} ITEMS)</span>
          <span>CYCLE 18000s</span>
        </div>

        <div className="sequence-items-track">
          {window.playlist && window.playlist.length > 0 ? (
            window.playlist.map((entry, idx) => {
              const isActive = !isSync && playback?.activeEntryId === entry.id;
              return (
                <div
                  key={entry.id || idx}
                  className={`track-item ${isActive ? 'is-active' : ''}`}
                >
                  {getIcon(entry.media?.type)}
                  <span>#{idx + 1} {entry.media?.type} ({entry.media?.duration_seconds}s)</span>
                </div>
              );
            })
          ) : (
            <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>
              No media configured in sequence track.
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
