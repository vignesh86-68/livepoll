import { useState, useEffect, useRef, useCallback } from 'react';
import { api } from '../api';

export function useWebSocket(code) {
  const [tally, setTally] = useState({ counts: {}, total: 0, seq: 0 });
  const [viewers, setViewers] = useState(0);
  const [status, setStatus] = useState('open');
  const [connected, setConnected] = useState(false);
  
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);
  const expectedSeq = useRef(null);
  const isComponentMounted = useRef(true);

  const connect = useCallback(() => {
    if (!code || !isComponentMounted.current) return;
    
    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/polls/${code}`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      reconnectAttempts.current = 0; // Reset attempts on successful connection
    };

    ws.onmessage = async (event) => {
      try {
        const data = JSON.parse(event.data);
        
        switch (data.type) {
          case 'snapshot':
            setTally(data.tally);
            setViewers(data.viewers);
            setStatus(data.status || 'open');
            expectedSeq.current = data.tally.seq + 1;
            break;
            
          case 'tally':
            // Sequence gap detection
            if (expectedSeq.current !== null && data.tally.seq > expectedSeq.current) {
              // Gap detected, fetch full state via HTTP
              try {
                const results = await api.getPollResults(code);
                setTally(results);
                expectedSeq.current = results.seq + 1;
              } catch (err) {
                console.error("Failed to resync tally", err);
              }
            } else {
              // In order, update normally
              setTally(data.tally);
              expectedSeq.current = data.tally.seq + 1;
            }
            break;
            
          case 'viewers':
            setViewers(data.viewers);
            break;
            
          case 'closed':
            setStatus('closed');
            if (data.tally) {
               setTally(data.tally);
            }
            break;
            
          default:
            console.warn('Unknown event type', data.type);
        }
      } catch (err) {
        console.error('WebSocket message parsing error', err);
      }
    };

    ws.onclose = () => {
      setConnected(false);
      wsRef.current = null;
      
      if (isComponentMounted.current && status !== 'closed') {
        // Exponential backoff with jitter
        const base = 1000;
        const max = 30000;
        const multiplier = Math.pow(2, reconnectAttempts.current);
        const jitter = Math.random() * 1000;
        const delay = Math.min(base * multiplier + jitter, max);
        
        reconnectAttempts.current += 1;
        
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, delay);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      // Let onclose handle reconnection
    };
  }, [code, status]);

  // Initial connection and cleanup
  useEffect(() => {
    isComponentMounted.current = true;
    connect();

    return () => {
      isComponentMounted.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  // Resync on tab focus
  useEffect(() => {
    const handleVisibilityChange = async () => {
      if (document.visibilityState === 'visible' && code && connected) {
         try {
            const results = await api.getPollResults(code);
            setTally(results);
            expectedSeq.current = results.seq + 1;
         } catch (err) {
            console.error('Failed to resync on focus', err);
         }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [code, connected]);

  return { tally, viewers, status, connected };
}
