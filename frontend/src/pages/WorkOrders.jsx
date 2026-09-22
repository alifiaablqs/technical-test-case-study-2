import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { Loading, ErrorState } from '../components/LoadingError';
import { Plus } from 'lucide-react';

const StatusBadge = ({ status }) => {
  const cls = `badge badge-${status.toLowerCase()}`;
  return <span className={cls}>{status}</span>;
};

const WorkOrders = () => {
  const [workOrders, setWorkOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // Create WO State
  const [isCreating, setIsCreating] = useState(false);
  const [products, setProducts] = useState([]);
  const [selectedProductId, setSelectedProductId] = useState('');
  const [quantity, setQuantity] = useState(1);
  const [activeBOM, setActiveBOM] = useState(null);
  const [createError, setCreateError] = useState(null);
  const navigate = useNavigate();

  const fetchWorkOrders = async () => {
    try {
      setLoading(true);
      const data = await api.get('/api/work-orders');
      setWorkOrders(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchProducts = async () => {
    try {
      const data = await api.get('/api/products');
      setProducts(data || []);
      if (data && data.length > 0) setSelectedProductId(data[0].id.toString());
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchWorkOrders();
  }, []);

  useEffect(() => {
    if (isCreating && products.length === 0) {
      fetchProducts();
    }
  }, [isCreating]);

  useEffect(() => {
    if (selectedProductId) {
      api.get(`/api/products/${selectedProductId}/bom`)
        .then(data => setActiveBOM(data))
        .catch(() => setActiveBOM(null));
    }
  }, [selectedProductId]);

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreateError(null);
    try {
      await api.post('/api/work-orders', {
        product_id: parseInt(selectedProductId),
        quantity: parseFloat(quantity)
      });
      setIsCreating(false);
      fetchWorkOrders();
    } catch (err) {
      setCreateError(err.message);
    }
  };

  if (loading && !isCreating) return <Loading />;
  if (error && !isCreating) return <ErrorState message={error} onRetry={fetchWorkOrders} />;

  if (isCreating) {
    return (
      <div style={{ maxWidth: '600px', margin: '0 auto' }}>
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Create Work Order</h2>
            <button className="btn btn-outline" onClick={() => setIsCreating(false)}>Cancel</button>
          </div>
          
          {createError && (
            <div className="alert alert-error">{createError}</div>
          )}

          <form onSubmit={handleCreate}>
            <div className="form-group">
              <label className="form-label">Product</label>
              <select 
                className="form-control"
                value={selectedProductId}
                onChange={(e) => setSelectedProductId(e.target.value)}
                required
              >
                {products.map(p => (
                  <option key={p.id} value={p.id}>{p.sku} - {p.name}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Quantity</label>
              <input 
                type="number" 
                className="form-control" 
                min="1" 
                step="1"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                required
              />
            </div>

            {activeBOM && quantity > 0 && (
              <div style={{ marginTop: '1.5rem', marginBottom: '1.5rem', backgroundColor: '#F8FAFC', padding: '1rem', borderRadius: '8px' }}>
                <h4 style={{ marginBottom: '1rem', fontSize: '0.875rem', fontWeight: 600, color: 'var(--color-text-muted)', textTransform: 'uppercase' }}>
                  BOM Explosion Preview
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {activeBOM.items.map(item => (
                    <div key={item.material_id} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.875rem' }}>
                      <span>{item.sku} {item.name}</span>
                      <span style={{ fontWeight: 500 }}>{item.quantity} × {quantity} = {item.quantity * quantity} {item.unit}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            {!activeBOM && selectedProductId && (
              <div className="alert alert-error" style={{ marginTop: '1rem' }}>
                This product does not have an active BOM. Cannot create work order.
              </div>
            )}

            <button type="submit" className="btn btn-primary" style={{ width: '100%' }} disabled={!activeBOM || quantity <= 0}>
              Submit Work Order
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.5rem' }}>
        <button className="btn btn-primary" onClick={() => setIsCreating(true)}>
          <Plus size={16} /> New Work Order
        </button>
      </div>

      <div className="card">
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Product</th>
                <th>Quantity</th>
                <th>Status</th>
                <th>Created At</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {workOrders.map(wo => (
                <tr key={wo.id}>
                  <td>WO-{wo.id.toString().padStart(4, '0')}</td>
                  <td>{wo.product_name || `Product ${wo.product_id}`}</td>
                  <td>{wo.quantity}</td>
                  <td><StatusBadge status={wo.status} /></td>
                  <td>{new Date(wo.created_at).toLocaleDateString()}</td>
                  <td>
                    <button 
                      className="btn btn-outline" 
                      style={{ padding: '0.25rem 0.75rem', fontSize: '0.75rem' }}
                      onClick={() => navigate(`/work-orders/${wo.id}`)}
                    >
                      View Detail
                    </button>
                  </td>
                </tr>
              ))}
              {workOrders.length === 0 && (
                <tr>
                  <td colSpan="6" style={{ textAlign: 'center', color: 'var(--color-text-muted)' }}>No work orders found</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default WorkOrders;
