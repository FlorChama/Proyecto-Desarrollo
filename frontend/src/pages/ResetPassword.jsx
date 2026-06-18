import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { resetPassword } from '../services/api'
import styles from './Auth.module.css'

export default function ResetPassword() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', new_password: '', confirm_password: '' })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    if (!form.email || !form.new_password || !form.confirm_password) {
      setError('Completá todos los campos')
      return
    }
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

  if (success) {
    return (
      <div className={styles.page}>
        <div className={styles.card}>
          <h1 className={styles.title}>TicketHub</h1>
          <div className={styles.successBox}>
            <p>Contraseña restablecida correctamente.</p>
            <button className={styles.submitBtn} onClick={() => navigate('/login')}>
              Ir al inicio de sesión
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>TicketHub</h1>
        <h2 className={styles.subtitle}>Restablecer contraseña</h2>
        <p className={styles.hint}>Ingresá tu email registrado y elegí una nueva contraseña.</p>

        <form onSubmit={handleSubmit} className={styles.form}>
          {error && <div className={styles.error}>{error}</div>}

          <div className={styles.field}>
            <label>Email</label>
            <input
              type="email"
              placeholder="tu@email.com"
              value={form.email}
              onChange={e => setForm({ ...form, email: e.target.value })}
            />
          </div>
          <div className={styles.field}>
            <label>Nueva contraseña</label>
            <input
              type="password"
              placeholder="••••••••"
              value={form.new_password}
              onChange={e => setForm({ ...form, new_password: e.target.value })}
            />
          </div>
          <div className={styles.field}>
            <label>Confirmar contraseña</label>
            <input
              type="password"
              placeholder="••••••••"
              value={form.confirm_password}
              onChange={e => setForm({ ...form, confirm_password: e.target.value })}
            />
          </div>

          <button type="submit" disabled={loading} className={styles.submitBtn}>
            {loading ? 'Restableciendo...' : 'Restablecer contraseña'}
          </button>

          <p className={styles.switchText}>
            <Link to="/login">Volver al inicio de sesión</Link>
          </p>
        </form>
      </div>
    </div>
  )
}
