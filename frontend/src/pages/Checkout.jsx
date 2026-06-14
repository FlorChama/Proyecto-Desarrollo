import { useState, useEffect } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { getEventById, buyTicket } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './Checkout.module.css'

// Mismo fallback de eventos mock
const MOCK_EVENTS = [
  { ID: 2,  title: 'Maroon 5',                         date: '2026-09-03', venue: 'Hipódromo de San Isidro, Buenos Aires',     image_url: '/events/internacional2.png', price: 200000, vipPrice: 400000,  category: 'internacional' },
  { ID: 3,  title: 'Rosalía — LUX Tour 2026',          date: '2026-08-01', venue: 'Movistar Arena, Buenos Aires',              image_url: '/events/internacional3.png', price: 200000, vipPrice: 650000,  category: 'internacional' },
  { ID: 4,  title: 'BTS World Tour — Arirang',         date: '2026-10-21', venue: 'Estadio Único de La Plata, Buenos Aires',   image_url: '/events/internacional1.png', price: 350000, vipPrice: 1000000, category: 'internacional' },
  { ID: 15, title: 'Aitana — Cuarto Azul World Tour',  date: '2026-10-21', venue: 'Movistar Arena, Buenos Aires',              image_url: '/events/internacional4.png', price: 80000,  vipPrice: 180000,  category: 'internacional' },
  { ID: 5,  title: 'Ciro y los Persas',                date: '2026-08-21', venue: 'Plaza de la Música, Córdoba',               image_url: '/events/nacional1.png',      price: 80000,                    category: 'nacional'       },
  { ID: 6,  title: 'Babasónicos',                      date: '2026-06-25', venue: 'Plaza de la Música, Córdoba',               image_url: '/events/nacional2.png',      price: 75000,                    category: 'nacional'       },
  { ID: 7,  title: 'Calamaro — Como Cantor',           date: '2026-07-08', venue: 'Movistar Arena, Buenos Aires',              image_url: '/events/nacional3.png',      price: 75000,  vipPrice: 110000,  category: 'nacional'       },
  { ID: 8,  title: 'El Mató a un Policía Motorizado',  date: '2026-09-19', venue: 'Club Ciudad de Buenos Aires',               image_url: '/events/nacional4.png',      price: 111000, vipPrice: 160000,  category: 'nacional'       },
  { ID: 9,  title: 'Divididos',                        date: '2026-07-04', venue: 'Estadio Obras, Buenos Aires',               image_url: '/events/nacional5.png',      price: 55000,  vipPrice: 100000,  category: 'nacional'       },
  { ID: 10, title: 'Maldita Felicidad',                date: '2026-06-28', venue: 'Teatro Ciudad de las Artes, Córdoba',       image_url: '/events/teatro1.png',        price: 45000,                    category: 'teatro'         },
  { ID: 11, title: 'Fito Páez — Segunda Gran Fiesta',  date: '2026-07-07', venue: 'Movistar Arena, Buenos Aires',              image_url: '/events/teatro2.png',        price: 60000,  vipPrice: 120000,  category: 'nacional'       },
  { ID: 12, title: 'Madama Butterfly',                 date: '2026-07-11', venue: 'Teatro Avenida, Buenos Aires',              image_url: '/events/teatro3.png',        price: 30000,                    category: 'teatro'         },
  { ID: 13, title: 'Darío Orsi — Hasta las Manos',     date: '2026-09-05', venue: 'Teatro Gran Rex, Buenos Aires',             image_url: '/events/standUp1.png',       price: 20000,                    category: 'standup'        },
  { ID: 14, title: 'Manuel Ángel Redondo',             date: '2026-08-30', venue: 'Studio Theater, Córdoba',                   image_url: '/events/standUp2.png',       price: 20000,                    category: 'standup'        },
]

const parseDate = (d) => new Date(d && d.length === 10 ? d + 'T12:00:00' : d)
const formatDateLabel = (d) => parseDate(d).toLocaleDateString('es-AR', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })
const formatPrice = (n) => n?.toLocaleString('es-AR', { style: 'currency', currency: 'ARS', minimumFractionDigits: 0 })

const PAYMENT_METHODS = [
  { id: 'credit', label: 'Tarjeta de crédito', icon: '💳', type: 'card' },
  { id: 'debit',  label: 'Tarjeta de débito',  icon: '💳', type: 'card' },
  { id: 'modo',   label: 'MODO',               icon: null, img: '/payments/modo.png',        type: 'qr' },
  { id: 'mp',     label: 'Mercado Pago',        icon: null, img: '/payments/mercadopago.png', type: 'qr' },
]

// QR placeholder geométrico (sin dependencia externa)
function FakeQR({ label }) {
  return (
    <div className={styles.qrBlock}>
      <div className={styles.qrBox}>
        <svg viewBox="0 0 100 100" width="160" height="160" xmlns="http://www.w3.org/2000/svg">
          {/* corners */}
          <rect x="5"  y="5"  width="28" height="28" fill="none" stroke="#111" strokeWidth="4"/>
          <rect x="10" y="10" width="18" height="18" fill="#111"/>
          <rect x="67" y="5"  width="28" height="28" fill="none" stroke="#111" strokeWidth="4"/>
          <rect x="72" y="10" width="18" height="18" fill="#111"/>
          <rect x="5"  y="67" width="28" height="28" fill="none" stroke="#111" strokeWidth="4"/>
          <rect x="10" y="72" width="18" height="18" fill="#111"/>
          {/* dots */}
          {[40,48,56,64].map(x => [40,48,56,64].map(y => (
            Math.random() > 0.4
              ? <rect key={`${x}-${y}`} x={x} y={y} width="6" height="6" fill="#111"/>
              : null
          )))}
          <rect x="40" y="5"  width="6" height="6" fill="#111"/>
          <rect x="50" y="5"  width="6" height="6" fill="#111"/>
          <rect x="40" y="15" width="6" height="6" fill="#111"/>
          <rect x="56" y="10" width="6" height="6" fill="#111"/>
          <rect x="5"  y="40" width="6" height="6" fill="#111"/>
          <rect x="5"  y="50" width="6" height="6" fill="#111"/>
          <rect x="15" y="44" width="6" height="6" fill="#111"/>
          <rect x="72" y="40" width="6" height="6" fill="#111"/>
          <rect x="84" y="48" width="6" height="6" fill="#111"/>
          <rect x="78" y="56" width="6" height="6" fill="#111"/>
        </svg>
      </div>
      <p className={styles.qrHint}>Escaneá el código con la app de <strong>{label}</strong> para pagar</p>
    </div>
  )
}

export default function Checkout() {
  const { id } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const { isAuthenticated, user } = useAuth()

  const stateData = location.state || {}

  const [event,   setEvent]   = useState(null)
  const [loading, setLoading] = useState(true)
  const [step,    setStep]    = useState(1)   // 1 = datos personales, 2 = pago, 3 = resultado
  const [success, setSuccess] = useState(false)
  const [failed,  setFailed]  = useState(false)
  const [buying,  setBuying]  = useState(false)

  // Paso 1 — datos personales
  const [personal, setPersonal] = useState({
    nombre: user?.name?.split(' ')[0] || '',
    apellido: user?.name?.split(' ')[1] || '',
    dni: '',
    email: user?.email || '',
    telefono: '',
  })
  const [personalError, setPersonalError] = useState('')

  // Paso 2 — pago
  const [payMethod, setPayMethod] = useState('')
  const [card, setCard] = useState({ number: '', expiry: '', cvv: '', holder: '' })
  const [payError, setPayError] = useState('')

  useEffect(() => {
    if (!isAuthenticated()) { navigate('/login', { state: { from: `/checkout/${id}` } }); return }
    getEventById(id)
      .then(res => setEvent(res.data.data))
      .catch(() => {
        const mock = MOCK_EVENTS.find(e => String(e.ID) === String(id))
        if (mock) setEvent(mock)
        else navigate('/')
      })
      .finally(() => setLoading(false))
  }, [id])

  const ticketType  = stateData.ticketType  || { name: 'General', price: event?.price || 0 }
  const quantity    = stateData.quantity    || 1
  const selectedDate = stateData.date       || event?.date || ''
  const total       = ticketType.price * quantity

  // ── Paso 1: validar y avanzar ──
  const handleContinue = (e) => {
    e.preventDefault()
    const { nombre, apellido, dni, email, telefono } = personal
    if (!nombre || !apellido || !dni || !email || !telefono) {
      setPersonalError('Completá todos los campos para continuar.')
      return
    }
    if (!/^\d{7,8}$/.test(dni)) {
      setPersonalError('El DNI debe tener 7 u 8 dígitos.')
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setPersonalError('Ingresá un email válido.')
      return
    }
    setPersonalError('')
    setStep(2)
  }

  // ── Paso 2: confirmar pago ──
  const handlePay = async (e) => {
    e.preventDefault()
    if (!payMethod) { setPayError('Seleccioná un método de pago.'); return }
    const method = PAYMENT_METHODS.find(m => m.id === payMethod)
    if (method?.type === 'card') {
      if (!card.number || !card.expiry || !card.cvv || !card.holder) {
        setPayError('Completá todos los datos de la tarjeta.')
        return
      }
    }
    setPayError('')
    setBuying(true)
    try {
      const methodMap = { credit: 'credit_card', debit: 'debit_card', modo: 'modo', mp: 'mercadopago' }
      await buyTicket(Number(id), methodMap[payMethod] || payMethod, total, ticketType.id || 'general')
      setSuccess(true)
      setStep(3)
    } catch {
      setFailed(true)
      setStep(3)
    } finally {
      setBuying(false)
    }
  }

  if (loading) return <div className={styles.loading}><div className={styles.spinner} /></div>
  if (!event) return null

  const selectedMethod = PAYMENT_METHODS.find(m => m.id === payMethod)

  // ── Resultado ──
  if (step === 3) return (
    <div className={styles.resultPage}>
      <div className={styles.resultCard}>
        <div className={`${styles.resultIcon} ${success ? styles.iconSuccess : styles.iconFail}`}>
          {success
            ? <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            : <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          }
        </div>
        <h1 className={styles.resultTitle}>
          {success ? 'El pago se ejecutó correctamente' : 'Operación fallida'}
        </h1>
        <p className={styles.resultSub}>
          {success
            ? `Tu entrada para ${event.title} fue confirmada. Te enviamos el detalle al email ${personal.email}.`
            : 'No pudimos procesar el pago. Verificá tus datos e intentá nuevamente.'
          }
        </p>
        <div className={styles.resultActions}>
          {success
            ? <button onClick={() => navigate('/mis-entradas')} className={styles.btnPrimary}>Ver mis entradas</button>
            : <button onClick={() => setStep(2)} className={styles.btnPrimary}>Reintentar</button>
          }
          <button onClick={() => navigate('/')} className={styles.btnSecondary}>Volver al inicio</button>
        </div>
      </div>
    </div>
  )

  return (
    <div className={styles.page}>
      {/* Panel izquierdo — resumen del evento */}
      <div className={styles.leftPanel}>
        <div className={styles.decorCircle1} />
        <div className={styles.decorCircle2} />
        <div className={styles.brandLogo} onClick={() => navigate('/')}>TicketHub</div>

        {event.image_url && (
          <img src={event.image_url} alt={event.title} className={styles.eventThumb} />
        )}
        <h2 className={styles.eventName}>{event.title}</h2>
        <p className={styles.eventVenue}>{event.venue}</p>
        {selectedDate && (
          <p className={styles.eventDate}>{formatDateLabel(selectedDate)}</p>
        )}

        <div className={styles.orderBox}>
          <div className={styles.orderRow}>
            <span>{quantity}× {ticketType.name}</span>
            <span>{formatPrice(ticketType.price)}</span>
          </div>
          <div className={styles.orderDivider} />
          <div className={`${styles.orderRow} ${styles.orderTotal}`}>
            <span>Total</span>
            <span>{formatPrice(total)}</span>
          </div>
        </div>
      </div>

      {/* Panel derecho — formulario */}
      <div className={styles.rightPanel}>
        {/* Steps indicator */}
        <div className={styles.steps}>
          <div className={`${styles.step} ${step >= 1 ? styles.stepActive : ''}`}>
            <div className={styles.stepCircle}>1</div>
            <span>Tus datos</span>
          </div>
          <div className={styles.stepLine} />
          <div className={`${styles.step} ${step >= 2 ? styles.stepActive : ''}`}>
            <div className={styles.stepCircle}>2</div>
            <span>Pago</span>
          </div>
        </div>

        {/* ── PASO 1: Datos personales ── */}
        {step === 1 && (
          <div className={styles.formCard}>
            <div className={styles.formHeader}>
              <h1>Tus datos</h1>
              <p>Completá la información para continuar con la compra</p>
            </div>
            <form onSubmit={handleContinue} className={styles.form}>
              {personalError && <div className={styles.errorBox}>{personalError}</div>}

              <div className={styles.row2}>
                <div className={styles.field}>
                  <label>Nombre</label>
                  <input
                    type="text"
                    value={personal.nombre}
                    onChange={e => setPersonal({ ...personal, nombre: e.target.value })}
                    placeholder="Juan"
                    required
                  />
                </div>
                <div className={styles.field}>
                  <label>Apellido</label>
                  <input
                    type="text"
                    value={personal.apellido}
                    onChange={e => setPersonal({ ...personal, apellido: e.target.value })}
                    placeholder="Pérez"
                    required
                  />
                </div>
              </div>

              <div className={styles.field}>
                <label>DNI</label>
                <input
                  type="text"
                  inputMode="numeric"
                  value={personal.dni}
                  onChange={e => setPersonal({ ...personal, dni: e.target.value.replace(/\D/g, '') })}
                  placeholder="12345678"
                  maxLength={8}
                  required
                />
              </div>

              <div className={styles.field}>
                <label>Email</label>
                <input
                  type="email"
                  value={personal.email}
                  onChange={e => setPersonal({ ...personal, email: e.target.value })}
                  placeholder="tu@email.com"
                  required
                />
              </div>

              <div className={styles.field}>
                <label>Número de teléfono</label>
                <input
                  type="tel"
                  value={personal.telefono}
                  onChange={e => setPersonal({ ...personal, telefono: e.target.value.replace(/\D/g, '') })}
                  placeholder="1123456789"
                  required
                />
              </div>

              <button type="submit" className={styles.submitBtn}>Continuar</button>
            </form>
          </div>
        )}

        {/* ── PASO 2: Método de pago ── */}
        {step === 2 && (
          <div className={styles.formCard}>
            <div className={styles.formHeader}>
              <h1>Método de pago</h1>
              <p>Elegí cómo querés abonar tu entrada</p>
            </div>
            <form onSubmit={handlePay} className={styles.form}>
              {payError && <div className={styles.errorBox}>{payError}</div>}

              {/* Selector de método */}
              <div className={styles.methodGrid}>
                {PAYMENT_METHODS.map(m => (
                  <button
                    key={m.id}
                    type="button"
                    className={`${styles.methodBtn} ${payMethod === m.id ? styles.methodSelected : ''}`}
                    onClick={() => setPayMethod(m.id)}
                  >
                    {m.img
                      ? <img src={m.img} alt={m.label} className={styles.methodImg} />
                      : <span className={styles.methodIcon}>{m.icon}</span>
                    }
                    <span className={styles.methodLabel}>{m.label}</span>
                  </button>
                ))}
              </div>

              {/* Formulario de tarjeta */}
              {selectedMethod?.type === 'card' && (
                <div className={styles.cardForm}>
                  <div className={styles.field}>
                    <label>Número de tarjeta</label>
                    <input
                      type="text"
                      inputMode="numeric"
                      value={card.number}
                      onChange={e => setCard({ ...card, number: e.target.value.replace(/\D/g, '').slice(0,16) })}
                      placeholder="0000 0000 0000 0000"
                      maxLength={16}
                    />
                  </div>
                  <div className={styles.field}>
                    <label>Titular de la tarjeta</label>
                    <input
                      type="text"
                      value={card.holder}
                      onChange={e => setCard({ ...card, holder: e.target.value })}
                      placeholder="JUAN PÉREZ"
                    />
                  </div>
                  <div className={styles.row2}>
                    <div className={styles.field}>
                      <label>Vencimiento</label>
                      <input
                        type="text"
                        value={card.expiry}
                        onChange={e => {
                          let v = e.target.value.replace(/\D/g, '').slice(0, 4)
                          if (v.length > 2) v = v.slice(0,2) + '/' + v.slice(2)
                          setCard({ ...card, expiry: v })
                        }}
                        placeholder="MM/AA"
                        maxLength={5}
                      />
                    </div>
                    <div className={styles.field}>
                      <label>CVV</label>
                      <input
                        type="text"
                        inputMode="numeric"
                        value={card.cvv}
                        onChange={e => setCard({ ...card, cvv: e.target.value.replace(/\D/g, '').slice(0,4) })}
                        placeholder="123"
                        maxLength={4}
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* QR para MODO / MP */}
              {selectedMethod?.type === 'qr' && (
                <FakeQR label={selectedMethod.label} />
              )}

              <div className={styles.payActions}>
                <button type="button" onClick={() => setStep(1)} className={styles.backBtn}>
                  ← Volver
                </button>
                <button type="submit" disabled={buying || !payMethod} className={styles.submitBtn}>
                  {buying ? 'Procesando...' : `Pagar ${formatPrice(total)}`}
                </button>
              </div>
            </form>
          </div>
        )}
      </div>
    </div>
  )
}
