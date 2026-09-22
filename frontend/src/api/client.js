const handleResponse = async (response) => {
  if (response.ok) {
    if (response.status === 204) return null;
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.indexOf("application/json") !== -1) {
      return response.json();
    }
    return null;
  }

  let errorMessage = 'An unexpected error occurred.';
  try {
    const errorData = await response.json();
    if (errorData.error) {
      errorMessage = errorData.error;
    }
  } catch (e) {
    // Cannot parse JSON error
  }

  let businessMessage = errorMessage;
  if (response.status === 400) {
    businessMessage = 'Input tidak valid. Harap periksa kembali data Anda.';
  } else if (response.status === 404) {
    businessMessage = 'Data tidak ditemukan.';
  } else if (response.status === 409) {
    businessMessage = 'Work Order belum dapat diselesaikan karena status atau stok tidak memenuhi ketentuan.';
  } else if (response.status >= 500) {
    businessMessage = 'Terjadi kesalahan pada server. Silakan coba beberapa saat lagi.';
  }

  throw new Error(businessMessage);
};

// Simple local storage mock for Work Orders list since the API might not exist
const getLocalWorkOrders = () => {
  try {
    return JSON.parse(localStorage.getItem('poc_work_orders') || '[]');
  } catch {
    return [];
  }
};

const saveLocalWorkOrder = (wo) => {
  const wos = getLocalWorkOrders();
  const existingIdx = wos.findIndex(w => w.id === wo.id);
  if (existingIdx >= 0) {
    wos[existingIdx] = wo;
  } else {
    wos.unshift(wo);
  }
  localStorage.setItem('poc_work_orders', JSON.stringify(wos));
};

export const api = {
  get: async (url) => {
    // Intercept to provide mocked list if API doesn't exist
    if (url === '/api/work-orders') {
      try {
        const response = await fetch(url);
        if (response.status === 404) {
          // Fallback to local storage if API doesn't exist
          return getLocalWorkOrders();
        }
        const data = await handleResponse(response);
        return data ? data.data : [];
      } catch (err) {
        return getLocalWorkOrders();
      }
    }

    const response = await fetch(url);
    const data = await handleResponse(response);
    
    // If fetching specific WO, update local cache
    if (url.startsWith('/api/work-orders/') && !url.includes('/complete') && data && data.data) {
      saveLocalWorkOrder(data.data);
    }
    
    return data ? data.data : null;
  },
  post: async (url, body) => {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    const data = await handleResponse(response);
    
    if (data && data.data && url === '/api/work-orders') {
      saveLocalWorkOrder(data.data);
    }
    return data ? data.data : null;
  },
};
