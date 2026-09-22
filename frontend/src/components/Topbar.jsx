import React from 'react';
import { Menu } from 'lucide-react';

const Topbar = ({ onMenuClick, title }) => {
  return (
    <header className="topbar">
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        <button className="mobile-menu-btn" onClick={onMenuClick}>
          <Menu size={24} />
        </button>
        <h1 className="topbar-title">{title}</h1>
      </div>
      <div>
        {/* Can add user profile or other actions here later */}
      </div>
    </header>
  );
};

export default Topbar;
