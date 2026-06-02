import { useNavigate } from 'react-router-dom'
import styles from './EventCard.module.css'

export default function EventCard({ event }) {
  const navigate = useNavigate()

  const formatDate = (dateStr) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('es-AR', { day: '2-digit', month: 'long', year: 'numeric' })
  }

  const formatTime = (dateStr) => {
    const date = new Date(dateStr)
    return date.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className={styles.card} onClick={() => navigate(`/eventos/${event.ID}`)}>
      <div className={styles.imageWrapper}>
        {event.image_url ? (
          <img src={event.image_url} alt={event.title} className={styles.image} />
        ) : (
          <div className={styles.imagePlaceholder}>
            <span>🎭</span>
          </div>
        )}
        <span className={styles.category}>{event.category || 'General'}</span>
        {event.available === 0 && <span className={styles.soldOut}>AGOTADO</span>}
      </div>

      <div className={styles.info}>
        <h3 className={styles.title}>{event.title}</h3>
        <p className={styles.venue}>📍 {event.venue}</p>
        <div className={styles.dateRow}>
          <span>📅 {formatDate(event.date)}</span>
          <span>🕐 {formatTime(event.date)}</span>
        </div>
        <div className={styles.footer}>
          <span className={styles.price}>
            {event.price === 0 ? 'Gratis' : `$${event.price.toLocaleString('es-AR')}`}
          </span>
          <span className={styles.available}>
            {event.available > 0 ? `${event.available} disponibles` : 'Sin disponibilidad'}
          </span>
        </div>
      </div>
    </div>
  )
}
