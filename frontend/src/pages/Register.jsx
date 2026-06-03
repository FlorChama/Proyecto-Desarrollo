import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { register as registerApi } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Auth.module.css'

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
    if (form.password.length < 6) {
      setError('La contraseña debe tener al menos 6 caracteres')
      return
    }
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
      {/* Panel visual izquierdo */}
      <div className={styles.leftPanel}>
        <div className={styles.decorCircle1} />
        <div className={styles.decorCircle2} />
        <div className={styles.brandLogo}>🎫</div>
        <div className={styles.brandName}>TicketHub</div>
        <p className={styles.brandTagline}>
          Creá tu cuenta y empezá a comprar entradas para los shows que más te gustan
        </p>
      </div>

      {/* Panel formulario */}
      <div className={styles.rightPanel}>
        <div className={styles.card}>
          <div className={styles.header}>
            <h1>Crear cuenta</h1>
            <p>Es gratis y solo toma un minuto</p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {error && <div className={styles.error}>⚠️ {error}</div>}

            <div className={styles.field}>
              <label>Nombre completo</label>
              <div className={styles.inputWrapper}>
                <span className={styles.inputIcon}>👤</span>
                <input
                  type="text"
                  name="name"
                  value={form.name}
                  onChange={handleChange}
                  placeholder="Tu nombre"
                  required
                />
              </div>
            </div>

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
                  placeholder="Mínimo 6 caracteres"
                  required
                  minLength={6}
                />
              </div>
            </div>

            <button type="submit" disabled={loading} className={styles.submitBtn}>
              {loading ? 'Creando cuenta...' : 'Crear cuenta gratis'}
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
