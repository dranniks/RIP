import Nav from "react-bootstrap/Nav";
import Navbar from "react-bootstrap/Navbar";
import { NavLink } from "react-router-dom";

export const AppNavbar = (): JSX.Element => (
  <div className="navbar-shell">
    <Navbar expand="md" className="app-navbar">
      <Navbar.Toggle aria-controls="main-navbar" />
      <Navbar.Collapse id="main-navbar">
        <Nav className="navbar-links">
          <NavLink to="/services" className={({ isActive }) => `navbar-link ${isActive ? "active" : ""}`}>
            Список услуг
          </NavLink>
        </Nav>
      </Navbar.Collapse>
    </Navbar>
  </div>
);
