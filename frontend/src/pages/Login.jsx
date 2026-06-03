import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { login as loginApi } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Auth.module.css'

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
      {/* Panel visual izquierdo */}
      <div className={styles.leftPanel}>
        <div className={styles.decorCircle1} />
        <div className={styles.decorCircle2} />
        <div className={styles.brandLogo}>🎫</div>
        <div className={styles.brandName}>TicketHub</div>
        <p className={styles.brandTagline}>
          Tu plataforma para los mejores conciertos, shows y eventos de Argentina
        </p>
      </div>

      {/* Panel formulario */}
      <div className={styles.rightPanel}>
        <div className={styles.card}>
          <div className={styles.header}>
            <h1>Bienvenido de vuelta</h1>
            <p>Ingresá a tu cuenta para ver tus entradas</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {error && <div className={styles.error}>⚠️ {error}</div>}

            <div className={styles.field}>
              <label>Email</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}>✉️</span>
                <input
                  type="email"
                  name="email"
                  value={form.email}
                  onChange={handleChange}
                  placeholder="tu@email.com"
                  required
                />
              </div>
            </div>

            <div className={styles.field}>
              <label>Contraseña</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}>🔒</span>
                <input
                  type="password"
                  name="password"
                  value={form.password}
                  onChange={handleChange}
                  placeholder="••••••••"
                  required
                />
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
