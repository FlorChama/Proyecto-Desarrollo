import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getEventById } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './EventDetail.module.css'

// Mismo fallback que Home.jsx — para cuando la DB está vacía
const MOCK_EVENTS = [
  {
    ID: 2,
    title: 'Maroon 5',
    description: 'La banda liderada por Adam Levine regresa a Argentina con su gira mundial presentando sus grandes éxitos a lo largo de más de dos décadas de carrera.',
    date: '2026-09-03', dates: ['2026-09-03'],
    venue: 'Hipódromo de San Isidro, Buenos Aires',
    capacity: 60000, available: 8500, category: 'internacional',
    image_url: '/events/internacional2.png', price: 200000, vipPrice: 400000, status: 'active',
  },
  {
    ID: 3,
    title: 'Rosalía — LUX Tour 2026',
    description: 'La artista española presenta su nuevo álbum en una producción audiovisual única que combina flamenco, electrónica y arte visual de vanguardia.',
    date: '2026-08-01', dates: ['2026-08-01', '2026-08-02', '2026-08-04', '2026-08-06'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 1200, category: 'internacional',
    image_url: '/events/internacional3.png', price: 200000, vipPrice: 650000, status: 'active',
  },
  {
    ID: 4,
    title: 'BTS World Tour — Arirang',
    description: 'La banda de K-pop más grande del mundo llega a Argentina con tres fechas en el Estadio Único de La Plata. Un show de escala mundial con producción de primer nivel.',
    date: '2026-10-21', dates: ['2026-10-21', '2026-10-23', '2026-10-24'],
    venue: 'Estadio Único de La Plata, Buenos Aires',
    capacity: 80000, available: 5200, category: 'internacional',
    image_url: '/events/internacional1.png', price: 350000, vipPrice: 1000000, status: 'active',
  },
  {
    ID: 15,
    title: 'Aitana — Cuarto Azul World Tour',
    description: 'La cantante española trae su gira mundial a Buenos Aires con el show más personal y emotivo de su carrera.',
    date: '2026-10-21', dates: ['2026-10-21', '2026-10-22'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 500, category: 'internacional',
    image_url: '/events/internacional4.png', price: 80000, vipPrice: 180000, status: 'active',
  },
  {
    ID: 5,
    title: 'Ciro y los Persas',
    description: 'El ex líder de Los Piojos regresa a Córdoba con su banda para un show histórico que repasa toda su trayectoria con canciones imprescindibles del rock nacional.',
    date: '2026-08-21', dates: ['2026-08-21'],
    venue: 'Plaza de la Música, Córdoba',
    capacity: 20000, available: 4000, category: 'nacional', noVip: true,
    image_url: '/events/nacional1.png', price: 80000, status: 'active',
  },
  {
    ID: 6,
    title: 'Babasónicos',
    description: 'La banda cordobesa presenta su nuevo show en la Plaza de la Música, repasando sus clásicos y canciones de su último disco.',
    date: '2026-06-25', dates: ['2026-06-25'],
    venue: 'Plaza de la Música, Córdoba',
    capacity: 15000, available: 2800, category: 'nacional', noVip: true,
    image_url: '/events/nacional2.png', price: 75000, status: 'active',
  },
  {
    ID: 7,
    title: 'Calamaro — Como Cantor',
    description: "Andrés Calamaro en un show íntimo donde repasa toda su trayectoria solista: desde los '90 hasta hoy, con arreglos únicos y el humor que lo caracteriza.",
    date: '2026-07-08', dates: ['2026-07-08'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 6000, category: 'nacional',
    image_url: '/events/nacional3.png', price: 75000, vipPrice: 110000, status: 'active',
  },
  {
    ID: 8,
    title: 'El Mató a un Policía Motorizado',
    description: 'El retorno de uno de los grupos más influyentes del indie argentino. Una noche de post-punk y canciones que marcaron una generación.',
    date: '2026-09-19', dates: ['2026-09-19'],
    venue: 'Club Ciudad de Buenos Aires',
    capacity: 5000, available: 800, category: 'nacional',
    image_url: '/events/nacional4.png', price: 111000, vipPrice: 160000, status: 'active',
  },
  {
    ID: 9,
    title: 'Divididos',
    description: 'Ricardo Mollo y Diego Arnedo vuelven a los escenarios con el clásico sonido de Divididos. Blues, rock y mucha energía en el Estadio Obras.',
    date: '2026-07-04', dates: ['2026-07-04'],
    venue: 'Estadio Obras, Buenos Aires',
    capacity: 8000, available: 1500, category: 'nacional',
    image_url: '/events/nacional5.png', price: 55000, vipPrice: 100000, status: 'active',
  },
  {
    ID: 10,
    title: 'Maldita Felicidad',
    description: 'La comedia protagonizada por Pablo Echarri, Paola Krum, Carlos Portaluppi e Inés Palombo. Una historia sobre el amor, la familia y los vínculos que nos definen.',
    date: '2026-06-28', dates: ['2026-06-28'],
    venue: 'Teatro Ciudad de las Artes, Córdoba',
    capacity: 1200, available: 340, category: 'teatro',
    image_url: '/events/teatro1.png', price: 45000, status: 'active',
  },
  {
    ID: 11,
    title: 'Fito Páez — Segunda Gran Fiesta ¡FA!',
    description: 'El rosarino regresa al Movistar Arena para una segunda noche de su show ¡FA!, un recorrido por 40 años de música argentina con invitados sorpresa.',
    date: '2026-07-07', dates: ['2026-07-07'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 2100, category: 'nacional',
    image_url: '/events/teatro2.png', price: 60000, vipPrice: 120000, status: 'active',
  },
  {
    ID: 12,
    title: 'Madama Butterfly',
    description: "La ópera de Puccini en una producción de Juventus Lyrica con dirección escénica de Ana D'Anna. Una de las obras más emotivas del repertorio lírico mundial.",
    date: '2026-07-11', dates: ['2026-07-11', '2026-07-12', '2026-07-18', '2026-07-19'],
    venue: 'Teatro Avenida, Buenos Aires',
    capacity: 900, available: 220, category: 'teatro',
    image_url: '/events/teatro3.png', price: 30000, status: 'active',
  },
  {
    ID: 13,
    title: 'Darío Orsi — Hasta las Manos',
    description: 'El nuevo show de stand up de Darío Orsi: observaciones sobre la vida cotidiana, relaciones y situaciones absurdas con el sello inconfundible del comediante argentino.',
    date: '2026-09-05', dates: ['2026-09-05'],
    venue: 'Teatro Gran Rex, Buenos Aires',
    capacity: 1200, available: 380, category: 'standup',
    image_url: '/events/standUp1.png', price: 20000, status: 'active',
  },
  {
    ID: 14,
    title: 'Manuel Ángel Redondo — Adulto Responsable',
    description: 'Stand up & Crowd Work en vivo en Córdoba. Una noche de improvisación, anécdotas reales y mucha risa con uno de los comediantes más frescos de la nueva escena.',
    date: '2026-08-30', dates: ['2026-08-30'],
    venue: 'Studio Theater, Córdoba',
    capacity: 500, available: 140, category: 'standup',
    image_url: '/events/standUp2.png', price: 20000, status: 'active',
  },
]

const isMusicCategory = (cat) => cat === 'nacional' || cat === 'internacional'

const TICKET_TYPES = (event) => {
  if (isMusicCategory(event?.category) && !event?.noVip) {
    return [
      { id: 'general', name: 'General', price: event.price },
      { id: 'vip',     name: 'VIP',     price: event.vipPrice ?? Math.round(event.price * 2) },
    ]
  }
  return [
    { id: 'general', name: 'General', price: event?.price },
  ]
}

const parseDate = (d) => new Date(d && d.length === 10 ? d + 'T12:00:00' : d)

const formatDateLabel = (d) =>
  parseDate(d).toLocaleDateString('es-AR', { weekday: 'short', day: '2-digit', month: 'long', year: 'numeric' }) + ' — 21:00 hs'

const formatPrice = (n) =>
  n?.toLocaleString('es-AR', { style: 'currency', currency: 'ARS', minimumFractionDigits: 2 })

export default function EventDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()

  const [event,         setEvent]         = useState(null)
  const [loading,       setLoading]       = useState(true)
  const [selectedDate,  setSelectedDate]  = useState('')
  const [selectedType,  setSelectedType]  = useState('')
  const [quantity,      setQuantity]      = useState(1)

  useEffect(() => {
    getEventById(id)
      .then(res => {
        const data = res.data.data
        if (data) {
          setEvent(data)
          const dates = data.dates || (data.date ? [data.date] : [])
          if (dates.length === 1) setSelectedDate(dates[0])
        }
      })
      .catch(() => {
        // Fallback a mock si la DB no tiene el evento
        const mock = MOCK_EVENTS.find(e => String(e.ID) === String(id))
        if (mock) {
          setEvent(mock)
          if (mock.dates?.length === 1) setSelectedDate(mock.dates[0])
        } else {
          navigate('/')
        }
      })
      .finally(() => setLoading(false))
  }, [id])

  const ticketTypes = TICKET_TYPES(event)
  const currentType = ticketTypes.find(t => t.id === selectedType)
  const canBuy = selectedDate && selectedType

  const handleBuy = () => {
    if (!isAuthenticated()) { navigate('/login'); return }
    navigate(`/checkout/${id}`, {
      state: { date: selectedDate, ticketType: currentType, quantity },
    })
  }

  if (loading) return (
    <div className={styles.loading}>
      <div className={styles.spinner} />
    </div>
  )
  if (!event) return null

  const allDates = event.dates || (event.date ? [event.date] : [])

  return (
    <div className={styles.page}>

      {/* ── Botones volver ── */}
      <div className={styles.backBar}>
        <button onClick={() => navigate(-1)} className={styles.backBtn}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          Volver
        </button>
        <button onClick={() => navigate('/')} className={styles.backBtn}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
          Inicio
        </button>
      </div>

      {/* ── Hero imagen ── */}
      <div className={styles.hero}>
        {event.image_url
          ? <img src={event.image_url} alt={event.title} className={styles.heroImg} />
          : <div className={styles.heroFallback} />
        }
        <div className={styles.heroOverlay} />
      </div>

      {/* ── Barra de compra ── */}
      <div className={styles.buyBar}>
        <div className={styles.buyBarInner}>

          {/* Fecha */}
          <div className={styles.selectWrap}>
            <select
              className={styles.select}
              value={selectedDate}
              onChange={e => setSelectedDate(e.target.value)}
            >
              {allDates.length > 1 && <option value="">Seleccioná una fecha</option>}
              {allDates.map(d => (
                <option key={d} value={d}>{formatDateLabel(d)}</option>
              ))}
            </select>
          </div>

          {/* Tipo de entrada */}
          <div className={styles.selectWrap}>
            <select
              className={styles.select}
              value={selectedType}
              onChange={e => setSelectedType(e.target.value)}
            >
              <option value="">Tipo de entrada</option>
              {ticketTypes.map(t => (
                <option key={t.id} value={t.id}>{t.name}</option>
              ))}
            </select>
          </div>

          {/* Precio */}
          <div className={styles.priceDisplay}>
            {currentType
              ? <><span className={styles.priceLabel}>{currentType.name}</span><span className={styles.priceValue}>{formatPrice(currentType.price)}</span></>
              : <span className={styles.pricePlaceholder}>— Seleccioná tipo —</span>
            }
          </div>

          {/* Cantidad */}
          <div className={styles.selectWrap}>
            <select
              className={styles.select}
              value={quantity}
              onChange={e => setQuantity(Number(e.target.value))}
            >
              {[1,2,3,4,5,6].map(n => (
                <option key={n} value={n}>Cantidad: {n}</option>
              ))}
            </select>
          </div>

          {/* Botón comprar */}
          <button
            className={`${styles.buyBtn} ${!canBuy ? styles.buyBtnDisabled : ''}`}
            onClick={handleBuy}
            disabled={!canBuy}
          >
            Comprar
          </button>
        </div>
      </div>

      {/* ── Descripción ── */}
      <div className={styles.content}>

        {event.description && (
          <div className={styles.descSection}>
            <p className={styles.descText}>{event.description}</p>
          </div>
        )}

        <div className={styles.divider} />

        {/* ── Métodos de pago ── */}
        <div className={styles.paySection}>
          <h2 className={styles.payTitle}>Medios de pago</h2>
          <div className={styles.payGrid}>
            {[
              { id: 'visa-credito',  label: 'Visa Crédito',    img: '/payments/visa.png' },
              { id: 'visa-debito',   label: 'Visa Débito',     img: '/payments/visa.png' },
              { id: 'mastercard',    label: 'Mastercard',      img: '/payments/mastercard.png' },
              { id: 'master-debit',  label: 'Master Debit',    img: '/payments/mastercard.png' },
              { id: 'amex',          label: 'Amex',            img: '/payments/amex.svg' },
              { id: 'modo',          label: 'MODO',            img: '/payments/modo.png' },
              { id: 'mercadopago',   label: 'Mercado Pago',    img: '/payments/mercadopago.png' },
            ].map(m => (
              <div key={m.id} className={styles.payItem}>
                <div className={styles.payImgWrap}>
                  {/* Cuando subas las imágenes se van a mostrar acá */}
                  <img
                    src={m.img}
                    alt={m.label}
                    className={styles.payImg}
                    onError={e => { e.currentTarget.style.display = 'none'; e.currentTarget.nextSibling.style.display = 'flex' }}
                  />
                  <div className={styles.payImgFallback} style={{ display: 'none' }}>
                    <span>{m.label.charAt(0)}</span>
                  </div>
                </div>
                <span className={styles.payLabel}>{m.label}</span>
              </div>
            ))}
          </div>
        </div>

        <div className={styles.divider} />

        {/* ── Verificación de compra ── */}
        <div className={styles.verSection}>
          <h2 className={styles.verTitle}>VERIFICACIÓN DE COMPRA</h2>
          {/* El contenido de la sección se completará próximamente */}
        </div>

      </div>
    </div>
  )
}
