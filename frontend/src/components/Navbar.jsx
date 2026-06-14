import { useState, useRef, useEffect } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import styles from './Navbar.module.css'

const CATEGORIES = [
  {
    label: 'Música',
    value: 'musica',
    children: [
      { label: 'Internacionales', value: 'internacional' },
      { label: 'Nacionales',      value: 'nacional' },
    ],
  },
  { label: 'Teatro', value: 'teatro' },
  { label: 'Stand up',       value: 'standup' },
]

export default function Navbar() {
  const { user, logout, isAuthenticated, isAdmin } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [dropOpen, setDropOpen] = useState(false)
  const dropRef = useRef(null)

  const handleLogout = () => { logout(); navigate('/') }
  const isActive = (path) => location.pathname === path

  // Cerrar dropdown al hacer click fuera
  useEffect(() => {
    const handler = (e) => { if (dropRef.current && !dropRef.current.contains(e.target)) setDropOpen(false) }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const goCategory = (value) => {
    setDropOpen(false)
    navigate(`/?cat=${value}`)
  }

  return (
    <nav className={styles.navbar}>
      {/* Logo — izquierda */}
      <Link to="/" className={styles.logo}>
        <span className={styles.logoText}>TicketHub</span>
      </Link>

      {/* Todo el resto — derecha */}
      <div className={styles.right}>

        {/* Mis Entradas / Gestión (solo autenticados) */}
        {isAuthenticated() && !isAdmin() && (
          <Link to="/mis-entradas" className={`${styles.navLink} ${isActive('/mis-entradas') ? styles.active : ''}`}>
            Mis Entradas
          </Link>
        )}
        {isAdmin() && (
          <Link to="/admin" className={`${styles.navLink} ${isActive('/admin') ? styles.active : ''}`}>
            Gestión de Eventos
          </Link>
        )}

        {/* Dropdown Eventos */}
        <div className={styles.dropWrap} ref={dropRef}>
          <div className={`${styles.navLink} ${styles.dropTrigger} ${location.search.includes('cat=') || isActive('/') ? styles.active : ''}`}>
            {/* Texto "Eventos" navega directo al home */}
            <span className={styles.dropLabel} onClick={() => { setDropOpen(false); navigate('/') }}>
              Eventos
            </span>
            {/* Chevron abre/cierra el dropdown */}
            <button className={styles.chevronBtn} onClick={() => setDropOpen(o => !o)}>
              <svg className={`${styles.chevron} ${dropOpen ? styles.chevronUp : ''}`} width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
            </button>
          </div>

          {dropOpen && (
            <div className={styles.dropdown}>
              <button className={styles.dropItem} onClick={() => { setDropOpen(false); navigate('/') }}>
                Todos los eventos
              </button>
              <div className={styles.dropDivider} />
              {CATEGORIES.map(cat => (
                cat.children ? (
                  <div key={cat.label}>
                    <button className={styles.dropItem} onClick={() => goCategory(cat.value)}>
                      {cat.label}
                    </button>
                    {cat.children.map(sub => (
                      <button key={sub.value} className={`${styles.dropItem} ${styles.dropSubItem}`} onClick={() => goCategory(sub.value)}>
                        {sub.label}
                      </button>
                    ))}
                  </div>
                ) : (
                  <button key={cat.value} className={styles.dropItem} onClick={() => goCategory(cat.value)}>
                    {cat.label}
                  </button>
                )
              ))}
            </div>
          )}
        </div>

        {/* Carrito */}
        <button className={styles.cartBtn} onClick={() => navigate('/checkout')} title="Carrito">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
            <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
          </svg>
        </button>

        {/* Auth */}
        <div className={styles.authSection}>
          {isAuthenticated() ? (
            <div className={styles.userMenu}>
              {isAdmin() && <span className={styles.adminBadge}>Admin</span>}
              <span className={styles.userName}>Hola, {user?.name?.split(' ')[0]}</span>
              <button onClick={handleLogout} className={styles.logoutBtn}>Cerrar sesión</button>
            </div>
          ) : (
            <div className={styles.authButtons}>
              <Link to="/login" className={styles.loginBtn}>Iniciar sesión</Link>
              <Link to="/register" className={styles.registerBtn}>Registrarse</Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  )
}
