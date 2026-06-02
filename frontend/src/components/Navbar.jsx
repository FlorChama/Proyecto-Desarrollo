import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import styles from './Navbar.module.css'

export default function Navbar() {
  const { user, logout, isAuthenticated, isAdmin } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <nav className={styles.navbar}>
      <Link to="/" className={styles.logo}>
        <span className={styles.logoIcon}>🎫</span>
        <span className={styles.logoText}>TicketHub</span>
      </Link>

      <div className={styles.navLinks}>
        <Link to="/" className={styles.navLink}>Eventos</Link>
        {isAuthenticated() && !isAdmin() && (
          <Link to="/mis-entradas" className={styles.navLink}>Mis Entradas</Link>
        )}
        {isAdmin() && (
          <Link to="/admin" className={styles.navLink}>Panel Admin</Link>
        )}
      </div>

      <div className={styles.authSection}>
        {isAuthenticated() ? (
          <div className={styles.userMenu}>
            <span className={styles.userName}>Hola, {user?.name}</span>
            <button onClick={handleLogout} className={styles.logoutBtn}>Cerrar sesión</button>
          </div>
        ) : (
          <div className={styles.authButtons}>
            <Link to="/login" className={styles.loginBtn}>Iniciar sesión</Link>
            <Link to="/register" className={styles.registerBtn}>Registrarse</Link>
          </div>
        )}
      </div>
    </nav>
  )
}
