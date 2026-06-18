import { useState, useEffect, useRef } from 'react'
import { getAdminEvents, createEvent, updateEvent, deleteEvent, getEventReport, uploadEventImage } from '../services/api'
import styles from './AdminPanel.module.css'

const emptyForm = {
  title: '', description: '', date: '', duration: '',
  venue: '', capacity: '', category: '', image_url: '', price: '', vip_price: ''
}

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
  const [imagePreview, setImagePreview] = useState('')
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef(null)

  const fetchEvents = async () => {
    setLoading(true)
    try {
      const res = await getAdminEvents()
      setEvents(res.data.data || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchEvents() }, [])

  const handleEdit = (event) => {
    setEditingEvent(event)
    const dateStr = new Date(event.date).toISOString().slice(0, 16)
    setForm({
      title: event.title, description: event.description, date: dateStr,
      duration: event.duration, venue: event.venue, capacity: event.capacity,
      category: event.category, image_url: event.image_url,
      price: event.price, vip_price: event.vip_price || ''
    })
    setImagePreview(event.image_url || '')
    setShowForm(true)
    setFormError('')
  }

  const handleNew = () => {
    setEditingEvent(null)
    setForm(emptyForm)
    setImagePreview('')
    setShowForm(true)
    setFormError('')
  }

  const handleImageChange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    setImagePreview(URL.createObjectURL(file))
    setUploading(true)
    try {
      const res = await uploadEventImage(file)
      setForm(f => ({ ...f, image_url: res.data.data.url }))
    } catch {
      setFormError('Error al subir la imagen')
    } finally {
      setUploading(false)
    }
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
      const payload = {
        ...form,
        capacity: Number(form.capacity),
        duration: Number(form.duration),
        price: Number(form.price),
        vip_price: Number(form.vip_price) || 0,
        date: new Date(form.date).toISOString()
      }
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
                <th>Precio General</th>
                <th>Precio VIP</th>
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
                  <td>{ev.vip_price > 0 ? `$${ev.vip_price?.toLocaleString('es-AR')}` : '—'}</td>
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
                  <label>Precio General *</label>
                  <input type="number" value={form.price} onChange={(e) => setForm({...form, price: e.target.value})} placeholder="0" min="0" />
                </div>
                <div className={styles.field}>
                  <label>Precio VIP</label>
                  <input type="number" value={form.vip_price} onChange={(e) => setForm({...form, vip_price: e.target.value})} placeholder="0 (opcional)" min="0" />
                </div>
                <div className={styles.field}>
                  <label>Categoría</label>
                  <select value={form.category} onChange={(e) => setForm({...form, category: e.target.value})}>
                    <option value="">Seleccionar</option>
                    {['Música', 'Teatro', 'Deporte', 'Arte', 'Tecnología', 'Otro'].map(c => <option key={c}>{c}</option>)}
                  </select>
                </div>

                {/* Imagen */}
                <div className={`${styles.field} ${styles.fullWidth}`}>
                  <label>Imagen del evento</label>
                  <div className={styles.imageUploadArea} onClick={() => fileInputRef.current?.click()}>
                    {imagePreview ? (
                      <img src={imagePreview} alt="preview" className={styles.imagePreview} />
                    ) : (
                      <div className={styles.imagePlaceholder}>
                        <span>+</span>
                        <p>Hacé clic para subir una imagen</p>
                      </div>
                    )}
                    {uploading && <div className={styles.uploadingOverlay}>Subiendo...</div>}
                  </div>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleImageChange}
                    style={{ display: 'none' }}
                  />
                  {imagePreview && (
                    <button type="button" className={styles.removeImageBtn} onClick={() => { setImagePreview(''); setForm(f => ({...f, image_url: ''})) }}>
                      Quitar imagen
                    </button>
                  )}
                </div>
              </div>

              <div className={`${styles.field} ${styles.fullWidth}`}>
                <label>Descripción</label>
                <textarea value={form.description} onChange={(e) => setForm({...form, description: e.target.value})} placeholder="Descripción del evento..." rows={3} />
              </div>
              <div className={styles.modalActions}>
                <button type="submit" disabled={saving || uploading} className={styles.saveBtn}>
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
            <button onClick={() => setReport(null)} className={styles.closeBtn}>Cerrar</button>
          </div>
        </div>
      )}
    </div>
  )
}
