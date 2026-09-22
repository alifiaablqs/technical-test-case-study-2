import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { Loading, ErrorState } from '../components/LoadingError';

const Inventory = () => {
  const [materials, setMaterials] = useState([]);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchInventory = async () => {
    try {
      setLoading(true);
      setError(null);
      const [mats, prods] = await Promise.all([
        api.get('/api/materials'),
        api.get('/api/products')
      ]);
      setMaterials(mats || []);
      setProducts(prods || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInventory();
  }, []);

  if (loading) return <Loading />;
  if (error) return <ErrorState message={error} onRetry={fetchInventory} />;

  return (
    <div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">Materials</h2>
        </div>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Name</th>
                <th>Unit</th>
                <th>On Hand</th>
                <th>Reserved</th>
                <th>Available</th>
              </tr>
            </thead>
            <tbody>
              {materials.map(m => (
                <tr key={m.id}>
                  <td>{m.sku}</td>
                  <td>{m.name}</td>
                  <td>{m.unit}</td>
                  <td>{m.on_hand}</td>
                  <td>{m.reserved}</td>
                  <td style={{ fontWeight: 600 }}>{m.on_hand - m.reserved}</td>
                </tr>
              ))}
              {materials.length === 0 && (
                <tr>
                  <td colSpan="6" style={{ textAlign: 'center', color: 'var(--color-text-muted)' }}>No materials found</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">Products (Finished Goods)</h2>
        </div>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Name</th>
                <th>Unit</th>
                <th>On Hand</th>
                <th>Reserved</th>
                <th>Available</th>
              </tr>
            </thead>
            <tbody>
              {products.map(p => (
                <tr key={p.id}>
                  <td>{p.sku}</td>
                  <td>{p.name}</td>
                  <td>{p.unit}</td>
                  <td>{p.on_hand}</td>
                  <td>{p.reserved}</td>
                  <td style={{ fontWeight: 600 }}>{p.on_hand - p.reserved}</td>
                </tr>
              ))}
              {products.length === 0 && (
                <tr>
                  <td colSpan="6" style={{ textAlign: 'center', color: 'var(--color-text-muted)' }}>No products found</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Inventory;
