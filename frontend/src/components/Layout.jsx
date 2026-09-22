import React, { useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import Topbar from './Topbar';

const getTitle = (pathname) => {
  if (pathname === '/') return 'Overview';
  if (pathname.startsWith('/inventory')) return 'Inventory';
  if (pathname.startsWith('/bom')) return 'Bill of Materials';
  if (pathname.startsWith('/work-orders')) return 'Work Orders';
  return 'Application';
};

const Layout = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const location = useLocation();

  return (
    <div className="app-container">
      <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />
      
      {/* Mobile overlay */}
      {isSidebarOpen && (
        <div 
          className="modal-overlay" 
          style={{ zIndex: 40 }} 
          onClick={() => setIsSidebarOpen(false)}
        />
      )}

      <main className="main-content">
        <Topbar 
          onMenuClick={() => setIsSidebarOpen(true)} 
          title={getTitle(location.pathname)} 
        />
        <div className="page-content">
          <Outlet />
        </div>
      </main>
    </div>
  );
};

export default Layout;
