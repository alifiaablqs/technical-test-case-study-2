import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { Loading, ErrorState } from '../components/LoadingError';
import ConfirmModal from '../components/ConfirmModal';
import { ArrowLeft, CheckCircle } from 'lucide-react';

const StatusBadge = ({ status }) => {
  const cls = `badge badge-${status.toLowerCase()}`;
  return <span className={cls}>{status}</span>;
};

const WorkOrderDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [wo, setWo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [completeError, setCompleteError] = useState(null);
  const [completeSuccess, setCompleteSuccess] = useState(false);

  const fetchDetail = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.get(`/api/work-orders/${id}`);
      setWo(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDetail();
  }, [id]);

  const handleComplete = async () => {
    setCompleteError(null);
    try {
      await api.post(`/api/work-orders/${id}/complete`);
      setIsConfirmOpen(false);
      setCompleteSuccess(true);
      fetchDetail(); // Refresh data
      // Hide success message after 3s
      setTimeout(() => setCompleteSuccess(false), 3000);
    } catch (err) {
      setIsConfirmOpen(false);
      setCompleteError(err.message);
    }
  };

  if (loading && !wo) return <Loading />;
  if (error) return <ErrorState message={error} onRetry={fetchDetail} />;
  if (!wo) return <ErrorState message="Work order not found" />;

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: '1.5rem', gap: '1rem' }}>
        <button className="btn btn-outline" onClick={() => navigate('/work-orders')} style={{ padding: '0.5rem' }}>
          <ArrowLeft size={16} />
        </button>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 600 }}>Work Order WO-{wo.id.toString().padStart(4, '0')}</h2>
      </div>

      {completeSuccess && (
        <div className="alert alert-success" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <CheckCircle size={20} /> Work Order completed successfully.
        </div>
      )}

      {completeError && (
        <div className="alert alert-error">
          {completeError}
        </div>
      )}

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">Details</h2>
          {wo.status === 'RESERVED' && (
            <button className="btn btn-primary" onClick={() => setIsConfirmOpen(true)}>
              Complete Work Order
            </button>
          )}
        </div>
        
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.5rem', marginBottom: '1rem' }}>
          <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>Product</p>
            <p style={{ fontWeight: 500 }}>{wo.product_sku} - {wo.product_name}</p>
          </div>
          <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>Quantity</p>
            <p style={{ fontWeight: 500 }}>{wo.quantity}</p>
          </div>
          <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>BOM Version</p>
            <p style={{ fontWeight: 500 }}>v{wo.bom_version}</p>
          </div>
          <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>Status</p>
            <p><StatusBadge status={wo.status} /></p>
          </div>
          <div>
            <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>Created At</p>
            <p style={{ fontWeight: 500 }}>{new Date(wo.created_at).toLocaleString()}</p>
          </div>
          {wo.completed_at && (
            <div>
              <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', marginBottom: '0.25rem' }}>Completed At</p>
              <p style={{ fontWeight: 500 }}>{new Date(wo.completed_at).toLocaleString()}</p>
            </div>
          )}
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">Material Requirements</h2>
        </div>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Material</th>
                <th>Required</th>
                <th>Reserved</th>
                <th>Issued</th>
              </tr>
            </thead>
            <tbody>
              {wo.items.map(item => (
                <tr key={item.id}>
                  <td>{item.material_sku} - {item.material_name}</td>
                  <td>{item.required_quantity} {item.material_unit}</td>
                  <td>{item.reserved_quantity} {item.material_unit}</td>
                  <td style={{ fontWeight: 600, color: wo.status === 'COMPLETED' ? 'var(--color-success)' : 'inherit' }}>
                    {item.issued_quantity} {item.material_unit}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <ConfirmModal 
        isOpen={isConfirmOpen}
        title="Complete Work Order"
        message={`Are you sure you want to complete WO-${wo.id.toString().padStart(4, '0')}? This will issue the reserved materials and increase finished goods stock.`}
        confirmText="Complete"
        onConfirm={handleComplete}
        onCancel={() => setIsConfirmOpen(false)}
      />
    </div>
  );
};

export default WorkOrderDetail;
