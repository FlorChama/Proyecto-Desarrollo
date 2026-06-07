import { useNavigate } from 'react-router-dom'
import styles from './EventCard.module.css'

export default function EventCard({ event, size = 'md', grid = false }) {
  const navigate = useNavigate()

  const formatDate = (d) =>
    new Date(d).toLocaleDateString('es-AR', { day: 'numeric', month: 'long', year: 'numeric' })

  return (
    <div
      className={`${styles.card} ${size === 'lg' ? styles.large : ''} ${grid ? styles.grid : ''}`}
      onClick={() => navigate(`/eventos/${event.ID}`)}
    >
      {/* Imagen */}
      <div className={styles.imageWrapper}>
        {event.image_url
          ? <img src={event.image_url} alt={event.title} className={styles.image} />
          : <div className={styles.placeholder}><span>{event.title?.charAt(0)}</span></div>
        }
        {event.available === 0 && <span className={styles.soldOut}>Agotado</span>}
      </div>

      {/* Info */}
      <div className={styles.info}>
        <div>
          <h3 className={styles.title}>{event.title}</h3>
          <p className={styles.date}>{formatDate(event.date)}</p>
          {event.venue && <p className={styles.venue}>{event.venue}</p>}
        </div>

        <div className={styles.footer} onClick={e => e.stopPropagation()}>
          <button
            className={styles.moreBtn}
            onClick={() => navigate(`/eventos/${event.ID}`)}
          >
            Más info
          </button>
          <button
            className={styles.buyBtn}
            onClick={() => navigate(`/eventos/${event.ID}`)}
          >
            Comprar
          </button>
        </div>
      </div>
    </div>
  )
}
