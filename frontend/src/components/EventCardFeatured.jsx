import { useNavigate } from 'react-router-dom'
import styles from './EventCardFeatured.module.css'

export default function EventCardFeatured({ event }) {
  const navigate = useNavigate()

  const formatDate = (d) => new Date(d).toLocaleDateString('es-AR', { day: '2-digit', month: 'long', year: 'numeric' })
  const formatTime = (d) => new Date(d).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })

  return (
    <div className={styles.card} onClick={() => navigate(`/eventos/${event.ID}`)}>
      <div className={styles.imageWrapper}>
        {event.image_url
          ? <img src={event.image_url} alt={event.title} className={styles.image} />
          : <div className={styles.placeholder}><span>{event.title?.charAt(0) || 'E'}</span></div>
        }
        <div className={styles.overlay} />
      </div>

      <div className={styles.info}>
        {event.category && <span className={styles.cat}>{event.category}</span>}
        <h2 className={styles.title}>{event.title}</h2>
        <p className={styles.venue}>{event.venue}</p>
        <p className={styles.date}>{formatDate(event.date)} · {formatTime(event.date)}</p>
        {event.description && <p className={styles.desc}>{event.description}</p>}

        <div className={styles.footer}>
          <div className={styles.footerLeft}>
            <span className={styles.price}>
              {event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}
            </span>
            <span className={`${styles.avail} ${event.available === 0 ? styles.availNo : event.available <= 10 ? styles.availLow : styles.availOk}`}>
              {event.available === 0 ? 'Sin disponibilidad' : event.available <= 10 ? `Solo ${event.available} entradas` : `${event.available} disponibles`}
            </span>
          </div>
          <button className={`${styles.btn} ${event.available === 0 ? styles.btnOff : ''}`}>
            {event.available === 0 ? 'Agotado' : 'Comprar entradas'}
          </button>
        </div>
      </div>
    </div>
  )
}
