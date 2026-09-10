import { useState, useEffect } from 'react';
import { fetchHealth } from '../api/client';

export function useServerTimeOffset() {
  const [offset, setOffset] = useState(0);
  const [isSynced, setIsSynced] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const measureOffset = async () => {
      try {
        const start = Date.now();
        const health = await fetchHealth();
        const end = Date.now();
        const latency = (end - start) / 2;
        const serverTime = new Date(health.timestamp).getTime();
        const calculatedOffset = (serverTime + latency) - end;

        if (isMounted) {
          setOffset(calculatedOffset);
          setIsSynced(true);
        }
      } catch (err) {
        console.warn('Could not measure server time offset, using local time:', err);
      }
    };

    measureOffset();
    const interval = setInterval(measureOffset, 60000); // re-sync clock every minute
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  return { offset, isSynced };
}
