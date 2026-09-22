import React, { useEffect, useState } from 'react';
import { api } from '../api/client';
import { Loading, ErrorState, EmptyState } from '../components/LoadingError';

const BOM = () => {
  const [products, setProducts] = useState([]);
  const [selectedProduct, setSelectedProduct] = useState('');
  const [bomDetail, setBomDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchProducts();
  }, []);

  const fetchProducts = async () => {
    try {
      setLoading(true);
      const data = await api.get('/api/products');
      setProducts(data || []);
      if (data && data.length > 0) {
        setSelectedProduct(data[0].id.toString());
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedProduct) {
      fetchBOM(selectedProduct);
    }
  }, [selectedProduct]);

  const fetchBOM = async (productId) => {
    try {
      setBomDetail(null);
      setError(null);
      const data = await api.get(`/api/products/${productId}/bom`);
      setBomDetail(data);
    } catch (err) {
      if (err.message === 'Data tidak ditemukan.') {
        // No BOM for this product
        setBomDetail(null);
      } else {
        setError(err.message);
      }
    }
  };

  if (loading && products.length === 0) return <Loading />;
  if (error && products.length === 0) return <ErrorState message={error} onRetry={fetchProducts} />;

  return (
    <div>
      <div className="card">
        <div className="form-group" style={{ maxWidth: '300px', marginBottom: 0 }}>
          <label className="form-label">Select Product</label>
          <select 
            className="form-control"
            value={selectedProduct} 
            onChange={(e) => setSelectedProduct(e.target.value)}
          >
            {products.map(p => (
              <option key={p.id} value={p.id}>{p.sku} - {p.name}</option>
            ))}
          </select>
        </div>
      </div>

      {bomDetail ? (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">
              BOM Version {bomDetail.version}
              <span className="badge" style={{ marginLeft: '1rem', backgroundColor: '#E0F2FE', color: '#0369A1' }}>Active</span>
            </h2>
          </div>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Material SKU</th>
                  <th>Material Name</th>
                  <th>Quantity Required</th>
                  <th>Unit</th>
                </tr>
              </thead>
              <tbody>
                {bomDetail.items.map(item => (
                  <tr key={item.material_id}>
                    <td>{item.sku}</td>
                    <td>{item.name}</td>
                    <td>{item.quantity}</td>
                    <td>{item.unit}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <EmptyState message="No active Bill of Materials found for this product." />
      )}
    </div>
  );
};

export default BOM;
