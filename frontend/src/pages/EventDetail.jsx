import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getEventById, buyTicket } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './EventDetail.module.css'

export default function EventDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [event, setEvent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [buying, setBuying] = useState(false)
  const [result, setResult] = useState(null) // { success, message, ticket }

  useEffect(() => {
    getEventById(id)
      .then((res) => setEvent(res.data.data))
      .catch(() => navigate('/'))
      .finally(() => setLoading(false))
  }, [id])

  const handleBuy = async () => {
    if (!isAuthenticated()) {
      navigate('/login')
      return
    }
    setBuying(true)
    try {
      const res = await buyTicket(Number(id))
      setResult({ success: true, message: '¡Compra exitosa! Revisá tu email.', ticket: res.data.data })
    } catch (err) {
      setResult({ success: false, message: err.response?.data?.error || 'Error al procesar la compra' })
    } finally {
      setBuying(false)
    }
  }

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('es-AR', {
      weekday: 'long', day: '2-digit', month: 'long', year: 'numeric'
    })
  }

  const formatTime = (dateStr) => {
    return new Date(dateStr).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
  }

  if (loading) return <div className={styles.loading}>Cargando evento...</div>
  if (!event) return null

  return (
    <div className={styles.page}>
      <div className={styles.hero}>
        {event.image_url ? (
          <img src={event.image_url} alt={event.title} className={styles.heroImg} />
        ) : (
          <div className={styles.heroPlaceholder}>🎭</div>
        )}
        <div className={styles.heroOverlay} />
        <div className={styles.heroContent}>
          <span className={styles.category}>{event.category || 'General'}</span>
          <h1 className={styles.title}>{event.title}</h1>
          <p className={styles.venue}>📍 {event.venue}</p>
        </div>
      </div>

      <div className={styles.body}>
        <div className={styles.info}>
          <div className={styles.infoGrid}>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Fecha</span>
              <span className={styles.infoValue}>{formatDate(event.date)}</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Horario</span>
              <span className={styles.infoValue}>{formatTime(event.date)}</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Duración</span>
              <span className={styles.infoValue}>{event.duration ? `${event.duration} min` : 'A confirmar'}</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Capacidad</span>
              <span className={styles.infoValue}>{event.capacity} personas</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Disponibles</span>
              <span className={`${styles.infoValue} ${event.available === 0 ? styles.soldOut : ''}`}>
                {event.available} entradas
              </span>
            </div>
          </div>

          {event.description && (
            <div className={styles.description}>
              <h2>Sobre el evento</h2>
              <p>{event.description}</p>
            </div>
          )}
        </div>

        <div className={styles.buyBox}>
          <div className={styles.priceTag}>
            {event.price === 0 ? 'Gratis' : `$${event.price.toLocaleString('es-AR')}`}
            <span className={styles.priceLabel}>por persona</span>
          </div>

          {result ? (
            <div className={`${styles.result} ${result.success ? styles.resultSuccess : styles.resultError}`}>
              <span>{result.success ? '🎉' : '❌'}</span>
              <p>{result.message}</p>
              {result.success && result.ticket?.qr_code && (
                <img src={result.ticket.qr_code} alt="QR Code" className={styles.qrImage} />
              )}
              {result.success && (
                <button onClick={() => navigate('/mis-entradas')} className={styles.myTicketsBtn}>
                  Ver mis entradas
                </button>
              )}
              {!result.success && (
                <button onClick={() => setResult(null)} className={styles.retryBtn}>Intentar de nuevo</button>
              )}
            </div>
          ) : (
            <button
              onClick={handleBuy}
              disabled={buying || event.available === 0}
              className={styles.buyBtn}
            >
              {buying ? 'Procesando...' : event.available === 0 ? 'Sin disponibilidad' : 'Comprar entrada'}
            </button>
          )}

          <button onClick={() => navigate(-1)} className={styles.backBtn}>← Volver</button>
        </div>
      </div>
    </div>
  )
}
