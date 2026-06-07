import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getEventById } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './EventDetail.module.css'

// Sectores simulados basados en el precio del evento
const getSectors = (basePrice) => [
  { id: 'campo', name: 'Campo', price: basePrice, available: Math.floor(Math.random() * 200) + 50 },
  { id: 'platea', name: 'Platea', price: Math.round(basePrice * 1.5), available: Math.floor(Math.random() * 100) + 20 },
  { id: 'vip', name: 'VIP', price: Math.round(basePrice * 2.5), available: Math.floor(Math.random() * 30) + 5 },
]

export default function EventDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [event, setEvent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [selectedSector, setSelectedSector] = useState(null)
  const [quantity, setQuantity] = useState(1)

  useEffect(() => {
    getEventById(id)
      .then(res => { setEvent(res.data.data); })
      .catch(() => navigate('/'))
      .finally(() => setLoading(false))
  }, [id])

  const handleBuy = () => {
    if (!isAuthenticated()) { navigate('/login'); return }
    navigate(`/checkout/${id}`, { state: { sector: selectedSector, quantity } })
  }

  const formatDate = (d) => new Date(d).toLocaleDateString('es-AR', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })
  const formatTime = (d) => new Date(d).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })

  if (loading) return <div className={styles.loading}><div className={styles.spinner} /></div>
  if (!event) return null

  const sectors = getSectors(event.price || 5000)
  const total = selectedSector ? selectedSector.price * quantity : 0

  return (
    <div className={styles.page}>
      {/* Hero de imagen */}
      <div className={styles.hero}>
        {event.image_url
          ? <img src={event.image_url} alt={event.title} className={styles.heroImg} />
          : <div className={styles.heroFallback} />
        }
        <div className={styles.heroOverlay} />
        <div className={styles.heroContent}>
          {event.category && <span className={styles.cat}>{event.category}</span>}
          <h1 className={styles.title}>{event.title}</h1>
          <p className={styles.venue}>{event.venue}</p>
        </div>
        <button onClick={() => navigate(-1)} className={styles.backBtn}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          Volver
        </button>
      </div>

      {/* Cuerpo en fondo claro */}
      <div className={styles.body}>
        <div className={styles.left}>
          {/* Info general */}
          <div className={styles.infoRow}>
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
              <span className={styles.infoLabel}>Capacidad total</span>
              <span className={styles.infoValue}>{event.capacity?.toLocaleString('es-AR')}</span>
            </div>
          </div>

          {event.description && (
            <div className={styles.desc}>
              <h2>Sobre el evento</h2>
              <p>{event.description}</p>
            </div>
          )}

          {/* Selección de sectores */}
          {event.available > 0 && (
            <div className={styles.sectors}>
              <h2>Seleccioná tu sector</h2>
              <div className={styles.sectorList}>
                {sectors.map(s => (
                  <div
                    key={s.id}
                    className={`${styles.sectorCard} ${selectedSector?.id === s.id ? styles.sectorSelected : ''} ${s.available === 0 ? styles.sectorOff : ''}`}
                    onClick={() => s.available > 0 && setSelectedSector(s)}
                  >
                    <div className={styles.sectorLeft}>
                      <div className={`${styles.sectorRadio} ${selectedSector?.id === s.id ? styles.radioOn : ''}`} />
                      <div>
                        <p className={styles.sectorName}>{s.name}</p>
                        <p className={styles.sectorAvail}>
                          {s.available === 0 ? 'Sin disponibilidad' : `${s.available} entradas`}
                        </p>
                      </div>
                    </div>
                    <span className={styles.sectorPrice}>
                      {s.price === 0 ? 'Gratis' : `$${s.price.toLocaleString('es-AR')}`}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Panel de compra */}
        <div className={styles.buyBox}>
          <div className={styles.buyBoxInner}>
            <h3>Tu orden</h3>

            {selectedSector ? (
              <>
                <div className={styles.orderRow}>
                  <span>Sector {selectedSector.name}</span>
                  <span>${selectedSector.price.toLocaleString('es-AR')}</span>
                </div>

                {/* Cantidad */}
                <div className={styles.quantityRow}>
                  <span>Cantidad</span>
                  <div className={styles.stepper}>
                    <button onClick={() => setQuantity(q => Math.max(1, q - 1))} disabled={quantity <= 1}>−</button>
                    <span>{quantity}</span>
                    <button onClick={() => setQuantity(q => Math.min(selectedSector.available, q + 1))} disabled={quantity >= selectedSector.available}>+</button>
                  </div>
                </div>

                <div className={styles.orderDivider} />
                <div className={`${styles.orderRow} ${styles.orderTotal}`}>
                  <span>Total</span>
                  <span>${total.toLocaleString('es-AR')}</span>
                </div>

                <button onClick={handleBuy} className={styles.buyBtn}>
                  Comprar {quantity > 1 ? `${quantity} entradas` : 'entrada'}
                </button>
              </>
            ) : (
              <div className={styles.noSector}>
                {event.available === 0
                  ? <p>Este evento no tiene entradas disponibles</p>
                  : <p>Seleccioná un sector para continuar</p>
                }
                {event.available > 0 && (
                  <span className={styles.availBadge}>{event.available} entradas disponibles</span>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
