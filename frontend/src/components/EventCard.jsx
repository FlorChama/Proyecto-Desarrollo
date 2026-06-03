import { useNavigate } from 'react-router-dom'
import styles from './EventCard.module.css'

export default function EventCard({ event }) {
  const navigate = useNavigate()

  const formatDate = (dateStr) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('es-AR', { day: '2-digit', month: 'short', year: 'numeric' })
  }

  const formatTime = (dateStr) => {
    return new Date(dateStr).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
  }

  const isLow = event.available > 0 && event.available <= 10

  return (
    <div className={styles.card} onClick={() => navigate(`/eventos/${event.ID}`)}>
      <div className={styles.imageWrapper}>
        {event.image_url ? (
          <img src={event.image_url} alt={event.title} className={styles.image} />
        ) : (
          <div className={styles.imagePlaceholder}>🎭</div>
        )}
        <div className={styles.imageOverlay} />
        {event.category && <span className={styles.category}>{event.category}</span>}
        {event.available === 0 && <span className={styles.soldOut}>AGOTADO</span>}
        <span className={styles.priceOverlay}>
          {event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}
        </span>
      </div>

      <div className={styles.info}>
        <h3 className={styles.title}>{event.title}</h3>
        <p className={styles.venue}>📍 {event.venue}</p>
        <div className={styles.dateRow}>
          <span>📅 {formatDate(event.date)}</span>
          <span>🕐 {formatTime(event.date)}</span>
        </div>
        <div className={styles.footer}>
          <span className={`${styles.available} ${isLow ? styles.availableLow : ''}`}>
            {event.available === 0
              ? 'Sin disponibilidad'
              : isLow
              ? `⚡ Solo ${event.available} entradas`
              : `${event.available} disponibles`}
          </span>
          {event.available > 0
            ? <span className={styles.buyBtn}>Comprar</span>
            : <span className={styles.buyBtnDisabled}>Agotado</span>
          }
        </div>
      </div>
    </div>
  )
}
