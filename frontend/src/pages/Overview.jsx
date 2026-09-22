import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { Loading, ErrorState } from '../components/LoadingError';
import { Package, ClipboardList, CheckCircle } from 'lucide-react';

const Overview = () => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchOverview = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Fetch required data in parallel
      const [materials, products, workOrders] = await Promise.all([
        api.get('/api/materials'),
        api.get('/api/products'),
        api.get('/api/work-orders') // Wait, there's no GET /api/work-orders endpoint in the API list provided by user.
        // Let's assume we can fetch them, wait the prompt didn't say GET /api/work-orders is available!
        // "API yang tersedia: GET /api/materials, GET /api/products, GET /api/products/:id/bom, POST /api/work-orders, GET /api/work-orders/:id, POST /api/work-orders/:id/complete"
        // Wait, "GET /api/work-orders" isn't listed, but "WORK ORDER LIST" requirement says: "Tampilkan Work Order: ID, Product, Quantity, Status, Created At, Action".
        // Let's assume GET /api/work-orders exists because otherwise we can't do WORK ORDER LIST. I will fetch it.
      ]);

      const reservedMaterialCount = materials?.reduce((acc, m) => acc + (m.reserved > 0 ? 1 : 0), 0) || 0;
      const activeWO = workOrders?.filter(wo => wo.status === 'RESERVED').length || 0;
      const completedWO = workOrders?.filter(wo => wo.status === 'COMPLETED').length || 0;

      setData({
        totalMaterials: materials?.length || 0,
        totalProducts: products?.length || 0,
        reservedMaterials: reservedMaterialCount,
        activeWorkOrders: activeWO,
        completedWorkOrders: completedWO
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOverview();
  }, []);

  if (loading) return <Loading />;
  if (error) return <ErrorState message={error} onRetry={fetchOverview} />;

  return (
    <div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem' }}>
        
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <div style={{ backgroundColor: '#E0F2FE', padding: '1rem', borderRadius: '8px', color: '#0369A1' }}>
              <Package size={24} />
            </div>
            <div>
              <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', fontWeight: 500 }}>Total Materials</p>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>{data.totalMaterials}</h2>
            </div>
          </div>
        </div>

        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <div style={{ backgroundColor: '#E0F2FE', padding: '1rem', borderRadius: '8px', color: '#0369A1' }}>
              <Package size={24} />
            </div>
            <div>
              <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', fontWeight: 500 }}>Total Products</p>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>{data.totalProducts}</h2>
            </div>
          </div>
        </div>

        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <div style={{ backgroundColor: '#FEF3C7', padding: '1rem', borderRadius: '8px', color: '#B45309' }}>
              <ClipboardList size={24} />
            </div>
            <div>
              <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', fontWeight: 500 }}>Active Work Orders</p>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>{data.activeWorkOrders}</h2>
            </div>
          </div>
        </div>

        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            <div style={{ backgroundColor: '#D1FAE5', padding: '1rem', borderRadius: '8px', color: '#047857' }}>
              <CheckCircle size={24} />
            </div>
            <div>
              <p style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem', fontWeight: 500 }}>Completed Work Orders</p>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>{data.completedWorkOrders}</h2>
            </div>
          </div>
        </div>

      </div>
    </div>
  );
};

export default Overview;
