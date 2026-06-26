import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { resetPassword } from '../services/api'
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

export default function ResetPassword() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', new_password: '', confirm_password: '' })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    if (form.new_password !== form.confirm_password) {
      setError('Las contraseñas no coinciden')
      return
    }
    if (form.new_password.length < 6) {
      setError('La contraseña debe tener al menos 6 caracteres')
      return
    }
    setLoading(true)
    try {
      await resetPassword({ email: form.email, new_password: form.new_password })
      setSuccess(true)
    } catch (err) {
      setError(err.response?.data?.error || 'Error al restablecer la contraseña')
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
        <div className={styles.brandName}>Recuperá tu cuenta</div>
        <p className={styles.brandTagline}>Ingresá tu email registrado y elegí una nueva contraseña para volver a acceder</p>
      </div>

      <div className={styles.rightPanel}>
        <div className={styles.card}>
          {success ? (
            <>
              <div className={styles.header}>
                <h1>¡Listo!</h1>
                <p>Tu contraseña fue restablecida correctamente.</p>
              </div>
              <button className={styles.submitBtn} onClick={() => navigate('/login')}>
                Ir al inicio de sesión
              </button>
            </>
          ) : (
            <>
              <div className={styles.header}>
                <h1>Restablecer contraseña</h1>
                <p>Completá los campos para crear una nueva contraseña</p>
              </div>

              <form onSubmit={handleSubmit} className={styles.form}>
                {error && <div className={styles.error}>{error}</div>}

                <div className={styles.field}>
                  <label>Email</label>
                  <div className={styles.inputWrapper}>
                    <span className={styles.inputIcon}><IconMail /></span>
                    <input
                      type="email"
                      placeholder="tu@email.com"
                      value={form.email}
                      onChange={e => setForm({ ...form, email: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div className={styles.field}>
                  <label>Nueva contraseña</label>
                  <div className={styles.inputWrapper}>
                    <span className={styles.inputIcon}><IconLock /></span>
                    <input
                      type="password"
                      placeholder="••••••••"
                      value={form.new_password}
                      onChange={e => setForm({ ...form, new_password: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div className={styles.field}>
                  <label>Confirmar contraseña</label>
                  <div className={styles.inputWrapper}>
                    <span className={styles.inputIcon}><IconLock /></span>
                    <input
                      type="password"
                      placeholder="••••••••"
                      value={form.confirm_password}
                      onChange={e => setForm({ ...form, confirm_password: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <button type="submit" disabled={loading} className={styles.submitBtn}>
                  {loading ? 'Restableciendo...' : 'Restablecer contraseña'}
                </button>
              </form>

              <p className={styles.switchAuth}>
                <Link to="/login">Volver al inicio de sesión</Link>
              </p>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
