import { useState } from 'react';
import logo from '../assets/company_logo.png';

const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <div className="navbar-logo">
          <img src={logo} alt="Logo" height={70} />
        </div>
       </div>
    </nav>
  );
};

export default Navbar;
