import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { changePassword } from '../services/api'
import styles from './Profile.module.css'

export default function Profile() {
  const { user } = useAuth()
  const [form, setForm] = useState({ current_password: '', new_password: '', confirm_password: '' })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    if (!form.current_password || !form.new_password || !form.confirm_password) {
      setError('Completá todos los campos')
      return
    }
    if (form.new_password !== form.confirm_password) {
      setError('Las contraseñas nuevas no coinciden')
      return
    }
    if (form.new_password.length < 6) {
      setError('La nueva contraseña debe tener al menos 6 caracteres')
      return
    }

    setLoading(true)
    try {
      await changePassword({ current_password: form.current_password, new_password: form.new_password })
      setSuccess('Contraseña actualizada correctamente')
      setForm({ current_password: '', new_password: '', confirm_password: '' })
    } catch (err) {
      setError(err.response?.data?.error || 'Error al cambiar la contraseña')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <h1>Mi Perfil</h1>

        <div className={styles.infoCard}>
          <div className={styles.infoRow}>
            <span className={styles.infoLabel}>Nombre</span>
            <span className={styles.infoValue}>{user?.name}</span>
          </div>
          <div className={styles.infoRow}>
            <span className={styles.infoLabel}>Email</span>
            <span className={styles.infoValue}>{user?.email}</span>
          </div>
          <div className={styles.infoRow}>
            <span className={styles.infoLabel}>Rol</span>
            <span className={styles.infoValue}>{user?.role === 'admin' ? 'Administrador' : 'Cliente'}</span>
          </div>
        </div>

        <div className={styles.section}>
          <h2>Cambiar contraseña</h2>
          <form onSubmit={handleSubmit} className={styles.form}>
            {error && <div className={styles.error}>{error}</div>}
            {success && <div className={styles.successMsg}>{success}</div>}

            <div className={styles.field}>
              <label>Contraseña actual</label>
              <input
                type="password"
                value={form.current_password}
                onChange={e => setForm({ ...form, current_password: e.target.value })}
                placeholder="••••••••"
              />
            </div>
            <div className={styles.field}>
              <label>Nueva contraseña</label>
              <input
                type="password"
                value={form.new_password}
                onChange={e => setForm({ ...form, new_password: e.target.value })}
                placeholder="••••••••"
              />
            </div>
            <div className={styles.field}>
              <label>Confirmar nueva contraseña</label>
              <input
                type="password"
                value={form.confirm_password}
                onChange={e => setForm({ ...form, confirm_password: e.target.value })}
                placeholder="••••••••"
              />
            </div>

            <button type="submit" disabled={loading} className={styles.submitBtn}>
              {loading ? 'Guardando...' : 'Cambiar contraseña'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
