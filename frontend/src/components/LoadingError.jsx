import React from 'react';
import { Loader2, AlertCircle } from 'lucide-react';

export const Loading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>
    <Loader2 className="animate-spin" size={32} style={{ animation: 'spin 1s linear infinite' }} />
    <span style={{ marginLeft: '1rem' }}>Loading data...</span>
    <style>{`@keyframes spin { 100% { transform: rotate(360deg); } }`}</style>
  </div>
);

export const ErrorState = ({ message, onRetry }) => (
  <div style={{ padding: '2rem', textAlign: 'center' }}>
    <AlertCircle size={48} color="var(--color-danger)" style={{ margin: '0 auto 1rem' }} />
    <h3 style={{ marginBottom: '0.5rem', color: 'var(--color-text-main)' }}>Error Loading Data</h3>
    <p style={{ color: 'var(--color-text-muted)', marginBottom: '1.5rem' }}>{message}</p>
    {onRetry && (
      <button className="btn btn-primary" onClick={onRetry}>Try Again</button>
    )}
  </div>
);

export const EmptyState = ({ message = "No data available" }) => (
  <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--color-text-muted)', backgroundColor: '#F8FAFC', borderRadius: '8px', border: '1px dashed var(--color-border)' }}>
    <p>{message}</p>
  </div>
);
