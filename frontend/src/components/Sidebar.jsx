import React from 'react';
import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Package, Boxes, ClipboardList, X } from 'lucide-react';

const Sidebar = ({ isOpen, onClose }) => {
  const navItems = [
    { name: 'Overview', path: '/', icon: <LayoutDashboard size={20} /> },
    { name: 'Inventory', path: '/inventory', icon: <Package size={20} /> },
    { name: 'BOM', path: '/bom', icon: <Boxes size={20} /> },
    { name: 'Work Orders', path: '/work-orders', icon: <ClipboardList size={20} /> },
  ];

  return (
    <aside className={`sidebar ${isOpen ? 'open' : ''}`}>
      <div className="sidebar-header">
        <div style={{ flex: 1 }}>Inventory Pro</div>
        {isOpen && (
          <button className="mobile-menu-btn" onClick={onClose} style={{ color: 'white' }}>
            <X size={24} />
          </button>
        )}
      </div>
      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <NavLink
            key={item.name}
            to={item.path}
            className={({ isActive }) => `nav-link ${isActive ? 'active' : ''}`}
            onClick={() => {
              if (window.innerWidth <= 768) onClose();
            }}
          >
            {item.icon}
            {item.name}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
};

export default Sidebar;
