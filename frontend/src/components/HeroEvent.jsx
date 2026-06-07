import { useNavigate } from 'react-router-dom'
import styles from './HeroEvent.module.css'

export default function HeroEvent({ event }) {
  const navigate = useNavigate()
  if (!event) return null

  const formatDate = (d) => new Date(d).toLocaleDateString('es-AR', { weekday: 'long', day: '2-digit', month: 'long' })
  const formatTime = (d) => new Date(d).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })

  return (
    <div className={styles.hero} onClick={() => navigate(`/eventos/${event.ID}`)}>
      {event.image_url
        ? <img src={event.image_url} alt={event.title} className={styles.bg} />
        : <div className={styles.bgFallback} />
      }
      <div className={styles.overlay} />

      <div className={styles.content}>
        <div className={styles.meta}>
          {event.category && <span className={styles.cat}>{event.category}</span>}
          <span className={styles.date}>{formatDate(event.date)} · {formatTime(event.date)}</span>
        </div>
        <h1 className={styles.title}>{event.title}</h1>
        <p className={styles.venue}>{event.venue}</p>
        {event.description && <p className={styles.desc}>{event.description}</p>}
        <div className={styles.actions}>
          <button className={styles.buyBtn} onClick={(e) => { e.stopPropagation(); navigate(`/eventos/${event.ID}`) }}>
            {event.available === 0 ? 'Sin disponibilidad' : 'Comprar entradas'}
          </button>
          <span className={styles.price}>
            {event.price === 0 ? 'Gratis' : `Desde $${Number(event.price).toLocaleString('es-AR')}`}
          </span>
          <span className={`${styles.avail} ${event.available === 0 ? styles.availNo : event.available <= 20 ? styles.availLow : ''}`}>
            {event.available === 0 ? 'Sin disponibilidad' : event.available <= 20 ? `Solo ${event.available} entradas` : `${event.available} entradas disponibles`}
          </span>
        </div>
      </div>
    </div>
  )
}
