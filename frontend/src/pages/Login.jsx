import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { login as loginApi } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Auth.module.css'

const IconMail = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>
  </svg>
)

const IconLock = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
  </svg>
)

export default function Login() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value })

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await loginApi(form)
      login(res.data.data.user, res.data.data.token)
      navigate('/')
    } catch (err) {
      setError(err.response?.data?.error || 'Email o contraseña incorrectos')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.leftPanel}>
        <div className={styles.decorCircle1} />
        <div className={styles.decorCircle2} />
        <div className={styles.brandLogo}>TicketHub</div>
        <div className={styles.brandName}>Bienvenido</div>
        <p className={styles.brandTagline}>Ingresá a tu cuenta para gestionar tus entradas y acceder a los mejores eventos</p>
      </div>

      <div className={styles.rightPanel}>
        <div className={styles.card}>
          <div className={styles.header}>
            <h1>Iniciar sesión</h1>
            <p>Ingresá tus datos para continuar</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {error && <div className={styles.error}>{error}</div>}

            <div className={styles.field}>
              <label>Email</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}><IconMail /></span>
                <input type="email" name="email" value={form.email} onChange={handleChange} placeholder="tu@email.com" required />
              </div>
            </div>

            <div className={styles.field}>
              <label>Contraseña</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}><IconLock /></span>
                <input type="password" name="password" value={form.password} onChange={handleChange} placeholder="••••••••" required />
              </div>
            </div>

            <button type="submit" disabled={loading} className={styles.submitBtn}>
              {loading ? 'Iniciando sesión...' : 'Iniciar sesión'}
            </button>
          </form>

          <p className={styles.switchAuth}>
            ¿No tenés cuenta? <Link to="/register">Registrate gratis</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
