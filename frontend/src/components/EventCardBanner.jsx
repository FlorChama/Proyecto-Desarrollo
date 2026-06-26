import { useNavigate } from 'react-router-dom'
import styles from './EventCardBanner.module.css'

export default function EventCardBanner({ event }) {
  const navigate = useNavigate()
  const formatDate = (d) => new Date(d).toLocaleDateString('es-AR', { day: '2-digit', month: 'short', year: 'numeric' })

  return (
    <div className={styles.card} onClick={() => navigate(`/eventos/${event.ID}`)}>
      {event.image_url
        ? <img src={event.image_url} alt={event.title} className={styles.img} />
        : <div className={styles.fallback}><span>{event.title?.charAt(0)}</span></div>
      }
      <div className={styles.overlay} />
      <div className={styles.content}>
        {event.category && <span className={styles.cat}>{event.category}</span>}
        <h3 className={styles.title}>{event.title}</h3>
        <p className={styles.date}>{formatDate(event.date)} · {event.venue}</p>
        <div className={styles.footer}>
          <span className={styles.price}>{event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}</span>
          <span className={`${styles.avail} ${event.available === 0 ? styles.no : event.available <= 15 ? styles.low : ''}`}>
            {event.available === 0 ? 'Agotado' : event.available <= 15 ? `Solo ${event.available}` : `${event.available} disp.`}
          </span>
        </div>
      </div>
    </div>
  )
}
