import { useState, useEffect } from 'react'
import { getEvents, createEvent, updateEvent, deleteEvent, getEventReport } from '../services/api'
import styles from './AdminPanel.module.css'

const emptyForm = { title: '', description: '', date: '', duration: '', venue: '', capacity: '', category: '', image_url: '', price: '' }

export default function AdminPanel() {
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingEvent, setEditingEvent] = useState(null)
  const [form, setForm] = useState(emptyForm)
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [report, setReport] = useState(null)
  const [message, setMessage] = useState(null)

  const fetchEvents = async () => {
    setLoading(true)
    try {
      const res = await getEvents()
      setEvents(res.data.data || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchEvents() }, [])

  const handleEdit = (event) => {
    setEditingEvent(event)
    const dateStr = new Date(event.date).toISOString().slice(0, 16)
    setForm({ title: event.title, description: event.description, date: dateStr, duration: event.duration, venue: event.venue, capacity: event.capacity, category: event.category, image_url: event.image_url, price: event.price })
    setShowForm(true)
    setFormError('')
  }

  const handleNew = () => {
    setEditingEvent(null)
    setForm(emptyForm)
    setShowForm(true)
    setFormError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!form.title || !form.date || !form.venue || !form.capacity || form.price === '') {
      setFormError('Completá todos los campos obligatorios')
      return
    }
    setSaving(true)
    setFormError('')
    try {
      const payload = { ...form, capacity: Number(form.capacity), duration: Number(form.duration), price: Number(form.price), date: new Date(form.date).toISOString() }
      if (editingEvent) {
        await updateEvent(editingEvent.ID, payload)
        setMessage({ type: 'success', text: 'Evento actualizado' })
      } else {
        await createEvent(payload)
        setMessage({ type: 'success', text: 'Evento creado exitosamente' })
      }
      setShowForm(false)
      fetchEvents()
    } catch (err) {
      setFormError(err.response?.data?.error || 'Error al guardar')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('¿Cancelar este evento?')) return
    try {
      await deleteEvent(id)
      setMessage({ type: 'success', text: 'Evento cancelado' })
      fetchEvents()
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || 'Error' })
    }
  }

  const handleReport = async (id) => {
    try {
      const res = await getEventReport(id)
      setReport(res.data.data)
    } catch {
      setMessage({ type: 'error', text: 'Error al obtener reporte' })
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.topBar}>
        <h1>Panel de Administración</h1>
        <button onClick={handleNew} className={styles.newBtn}>+ Nuevo Evento</button>
      </div>

      {message && (
        <div className={`${styles.message} ${message.type === 'success' ? styles.msgSuccess : styles.msgError}`}>
          {message.text}
          <button onClick={() => setMessage(null)}>✕</button>
        </div>
      )}

      {loading ? (
        <div className={styles.loading}>Cargando eventos...</div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Título</th>
                <th>Venue</th>
                <th>Fecha</th>
                <th>Cap.</th>
                <th>Dispon.</th>
                <th>Precio</th>
                <th>Estado</th>
                <th>Acciones</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev) => (
                <tr key={ev.ID}>
                  <td>{ev.title}</td>
                  <td>{ev.venue}</td>
                  <td>{new Date(ev.date).toLocaleDateString('es-AR')}</td>
                  <td>{ev.capacity}</td>
                  <td>{ev.available}</td>
                  <td>${ev.price?.toLocaleString('es-AR')}</td>
                  <td><span className={ev.status === 'active' ? styles.statusActive : styles.statusCancelled}>{ev.status}</span></td>
                  <td>
                    <div className={styles.actionBtns}>
                      <button onClick={() => handleEdit(ev)} className={styles.editBtn}>Editar</button>
                      <button onClick={() => handleReport(ev.ID)} className={styles.reportBtn}>Reporte</button>
                      <button onClick={() => handleDelete(ev.ID)} className={styles.deleteBtn}>Cancelar</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showForm && (
        <div className={styles.modalOverlay} onClick={(e) => e.target === e.currentTarget && setShowForm(false)}>
          <div className={styles.modal}>
            <h2>{editingEvent ? 'Editar Evento' : 'Nuevo Evento'}</h2>
            <form onSubmit={handleSubmit} className={styles.form}>
              {formError && <div className={styles.formError}>{formError}</div>}
              <div className={styles.formGrid}>
                <div className={styles.field}>
                  <label>Título *</label>
                  <input value={form.title} onChange={(e) => setForm({...form, title: e.target.value})} placeholder="Nombre del evento" />
                </div>
                <div className={styles.field}>
                  <label>Venue *</label>
                  <input value={form.venue} onChange={(e) => setForm({...form, venue: e.target.value})} placeholder="Lugar del evento" />
                </div>
                <div className={styles.field}>
                  <label>Fecha y hora *</label>
                  <input type="datetime-local" value={form.date} onChange={(e) => setForm({...form, date: e.target.value})} />
                </div>
                <div className={styles.field}>
                  <label>Duración (min)</label>
                  <input type="number" value={form.duration} onChange={(e) => setForm({...form, duration: e.target.value})} placeholder="120" />
                </div>
                <div className={styles.field}>
                  <label>Capacidad *</label>
                  <input type="number" value={form.capacity} onChange={(e) => setForm({...form, capacity: e.target.value})} placeholder="1000" min="1" />
                </div>
                <div className={styles.field}>
                  <label>Precio *</label>
                  <input type="number" value={form.price} onChange={(e) => setForm({...form, price: e.target.value})} placeholder="0" min="0" />
                </div>
                <div className={styles.field}>
                  <label>Categoría</label>
                  <select value={form.category} onChange={(e) => setForm({...form, category: e.target.value})}>
                    <option value="">Seleccionar</option>
                    {['Música', 'Teatro', 'Deporte', 'Arte', 'Tecnología', 'Otro'].map(c => <option key={c}>{c}</option>)}
                  </select>
                </div>
                <div className={styles.field}>
                  <label>URL Imagen</label>
                  <input value={form.image_url} onChange={(e) => setForm({...form, image_url: e.target.value})} placeholder="https://..." />
                </div>
              </div>
              <div className={`${styles.field} ${styles.fullWidth}`}>
                <label>Descripción</label>
                <textarea value={form.description} onChange={(e) => setForm({...form, description: e.target.value})} placeholder="Descripción del evento..." rows={3} />
              </div>
              <div className={styles.modalActions}>
                <button type="submit" disabled={saving} className={styles.saveBtn}>
                  {saving ? 'Guardando...' : 'Guardar'}
                </button>
                <button type="button" onClick={() => setShowForm(false)} className={styles.cancelModalBtn}>
                  Cancelar
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {report && (
        <div className={styles.modalOverlay} onClick={(e) => e.target === e.currentTarget && setReport(null)}>
          <div className={styles.modal}>
            <h2>Reporte: {report.title}</h2>
            <div className={styles.reportGrid}>
              <div className={styles.reportStat}><span>{report.total_capacity}</span><label>Capacidad total</label></div>
              <div className={styles.reportStat}><span>{report.total_sold}</span><label>Vendidas</label></div>
              <div className={styles.reportStat}><span>{report.total_cancelled}</span><label>Canceladas</label></div>
              <div className={styles.reportStat}><span>{report.available}</span><label>Disponibles</label></div>
            </div>
            {report.buyers?.length > 0 && (
              <div className={styles.buyerList}>
                <h3>Compradores ({report.buyers.length})</h3>
                <div className={styles.buyers}>
                  {report.buyers.map((b) => (
                    <div key={b.ID} className={styles.buyer}>
                      <span>{b.name}</span>
                      <span>{b.email}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <button onClick={() => setReport(null)} className={styles.closeBtn}>Cerrar</button>
          </div>
        </div>
      )}
    </div>
  )
}
