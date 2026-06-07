import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getEventById, buyTicket } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Checkout.module.css'

export default function Checkout() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [event, setEvent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [buying, setBuying] = useState(false)
  const [success, setSuccess] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!isAuthenticated()) { navigate('/login'); return }
    getEventById(id)
      .then(res => setEvent(res.data.data))
      .catch(() => navigate('/'))
      .finally(() => setLoading(false))
  }, [id])

  const handleConfirm = async () => {
    setBuying(true)
    setError('')
    try {
      const res = await buyTicket(Number(id))
      setSuccess(res.data.data)
    } catch (err) {
      setError(err.response?.data?.error || 'Error al procesar la compra')
    } finally {
      setBuying(false)
    }
  }

  const formatDate = (d) => new Date(d).toLocaleDateString('es-AR', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })
  const formatTime = (d) => new Date(d).toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })

  if (loading) return <div className={styles.loading}><div className={styles.spinner} /></div>
  if (!event) return null

  if (success) return (
    <div className={styles.successPage}>
      <div className={styles.successCard}>
        <div className={styles.successIconWrapper}>
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        </div>
        <h1 className={styles.successTitle}>Compra exitosa</h1>
        <p className={styles.successSub}>Tu entrada para <strong>{event.title}</strong> fue confirmada. Revisá tu email, te enviamos el código QR.</p>
        {success.qr_code && (
          <div className={styles.qrWrapper}>
            <img src={success.qr_code} alt="QR Code" className={styles.qr} />
            <p className={styles.qrLabel}>Presentá este código en el ingreso al evento</p>
          </div>
        )}
        <div className={styles.successActions}>
          <button onClick={() => navigate('/mis-entradas')} className={styles.btnPrimary}>Ver mis entradas</button>
          <button onClick={() => navigate('/')} className={styles.btnSecondary}>Volver al inicio</button>
        </div>
      </div>
    </div>
  )

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <div className={styles.left}>
          <h1 className={styles.pageTitle}>Resumen de compra</h1>

          <div className={styles.eventCard}>
            {event.image_url && <img src={event.image_url} alt={event.title} className={styles.eventImg} />}
            <div className={styles.eventInfo}>
              {event.category && <span className={styles.cat}>{event.category}</span>}
              <h2 className={styles.eventTitle}>{event.title}</h2>
              <p className={styles.eventMeta}>{event.venue}</p>
              <p className={styles.eventMeta}>{formatDate(event.date)} · {formatTime(event.date)}</p>
              {event.description && <p className={styles.eventDesc}>{event.description}</p>}
            </div>
          </div>

          <div className={styles.orderSummary}>
            <h3>Detalle de la orden</h3>
            <div className={styles.row}><span>1x Entrada general</span><span>{event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}</span></div>
            <div className={styles.divider} />
            <div className={`${styles.row} ${styles.total}`}><span>Total</span><span>{event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}</span></div>
          </div>
        </div>

        <div className={styles.right}>
          <div className={styles.confirmBox}>
            <h2>Confirmar compra</h2>
            <p className={styles.confirmDesc}>Al confirmar recibirás tu entrada con el código QR al email registrado.</p>

            <div className={styles.benefits}>
              <div className={styles.benefitItem}>Confirmación inmediata por email</div>
              <div className={styles.benefitItem}>Código QR para presentar en el evento</div>
              <div className={styles.benefitItem}>Podés transferir o cancelar desde Mis Entradas</div>
            </div>

            {error && <div className={styles.errorMsg}>{error}</div>}

            <div className={styles.priceBig}>
              {event.price === 0 ? 'Gratis' : `$${Number(event.price).toLocaleString('es-AR')}`}
            </div>

            <button onClick={handleConfirm} disabled={buying || event.available === 0} className={styles.confirmBtn}>
              {buying ? 'Procesando...' : event.available === 0 ? 'Sin disponibilidad' : 'Confirmar compra'}
            </button>
            <button onClick={() => navigate(`/eventos/${id}`)} className={styles.backBtn}>Volver al evento</button>
          </div>
        </div>
      </div>
    </div>
  )
}
