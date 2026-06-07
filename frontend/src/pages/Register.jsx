import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { register as registerApi } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Auth.module.css'

const IconUser = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
  </svg>
)

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

export default function Register() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value })

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    if (form.password.length < 6) { setError('La contraseña debe tener al menos 6 caracteres'); return }
    setLoading(true)
    try {
      const res = await registerApi(form)
      login(res.data.data.user, res.data.data.token)
      navigate('/')
    } catch (err) {
      setError(err.response?.data?.error || 'Error al crear la cuenta')
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
        <div className={styles.brandName}>Creá tu cuenta</div>
        <p className={styles.brandTagline}>Registrate y empezá a comprar entradas para los shows que más te gustan</p>
      </div>

      <div className={styles.rightPanel}>
        <div className={styles.card}>
          <div className={styles.header}>
            <h1>Crear cuenta</h1>
            <p>Completá tus datos para registrarte</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {error && <div className={styles.error}>{error}</div>}

            <div className={styles.field}>
              <label>Nombre completo</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}><IconUser /></span>
                <input type="text" name="name" value={form.name} onChange={handleChange} placeholder="Tu nombre" required />
              </div>
            </div>

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
                <input type="password" name="password" value={form.password} onChange={handleChange} placeholder="Mínimo 6 caracteres" required minLength={6} />
              </div>
            </div>

            <button type="submit" disabled={loading} className={styles.submitBtn}>
              {loading ? 'Creando cuenta...' : 'Crear cuenta'}
            </button>
          </form>

          <p className={styles.switchAuth}>
            ¿Ya tenés cuenta? <Link to="/login">Iniciá sesión</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
