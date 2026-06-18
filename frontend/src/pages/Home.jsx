import { useState, useEffect, useRef } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { getEvents } from '../services/api'
import EventCard from '../components/EventCard'
import styles from './Home.module.css'

// Eventos de prueba — se usan cuando la base de datos está vacía
// "dates" lista todas las fechas del evento; "date" es la fecha principal (para mostrar)
const MOCK_EVENTS = [
  {
    ID: 2,
    title: 'Maroon 5',
    description: 'La banda liderada por Adam Levine regresa a Argentina con su gira mundial.',
    date: '2026-09-03', dates: ['2026-09-03'],
    venue: 'Hipódromo de San Isidro, Buenos Aires',
    capacity: 60000, available: 8500, category: 'internacional',
    image_url: '/events/internacional2.png', price: 200000, vipPrice: 400000, status: 'active',
  },
  {
    ID: 3,
    title: 'Rosalía — LUX Tour 2026',
    description: 'La artista española presenta su nuevo álbum en una producción audiovisual única.',
    date: '2026-08-01', dates: ['2026-08-01', '2026-08-02', '2026-08-04', '2026-08-06'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 1200, category: 'internacional',
    image_url: '/events/internacional3.png', price: 200000, vipPrice: 650000, status: 'active',
  },
  {
    ID: 4,
    title: 'BTS World Tour — Arirang',
    description: 'La banda de K-pop más grande del mundo llega a Argentina con tres fechas en el Estadio Único de La Plata.',
    date: '2026-10-21', dates: ['2026-10-21', '2026-10-23', '2026-10-24'],
    venue: 'Estadio Único de La Plata, Buenos Aires',
    capacity: 80000, available: 5200, category: 'internacional',
    image_url: '/events/internacional1.png', price: 350000, vipPrice: 1000000, status: 'active',
  },
  {
    ID: 15,
    title: 'Aitana — Cuarto Azul World Tour',
    description: 'La cantante española trae su gira mundial a Buenos Aires.',
    date: '2026-10-21', dates: ['2026-10-21', '2026-10-22'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 500, category: 'internacional',
    image_url: '/events/internacional4.png', price: 80000, vipPrice: 180000, status: 'active',
  },
  {
    ID: 5,
    title: 'Ciro y los Persas',
    description: 'El ex líder de Los Piojos regresa a Córdoba con su banda para un show histórico.',
    date: '2026-08-21', dates: ['2026-08-21'],
    venue: 'Plaza de la Música, Córdoba',
    capacity: 20000, available: 4000, category: 'nacional', noVip: true,
    image_url: '/events/nacional1.png', price: 80000, status: 'active',
  },
  {
    ID: 6,
    title: 'Babasónicos',
    description: 'La banda cordobesa presenta su nuevo show en la Plaza de la Música.',
    date: '2026-06-25', dates: ['2026-06-25'],
    venue: 'Plaza de la Música, Córdoba',
    capacity: 15000, available: 2800, category: 'nacional', noVip: true,
    image_url: '/events/nacional2.png', price: 75000, status: 'active',
  },
  {
    ID: 7,
    title: 'Calamaro — Como Cantor',
    description: 'Andrés Calamaro en un show íntimo donde repasa toda su trayectoria solista.',
    date: '2026-07-08', dates: ['2026-07-08'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 6000, category: 'nacional',
    image_url: '/events/nacional3.png', price: 75000, vipPrice: 110000, status: 'active',
  },
  {
    ID: 8,
    title: 'El Mató a un Policía Motorizado',
    description: 'El retorno de uno de los grupos más influyentes del indie argentino.',
    date: '2026-09-19', dates: ['2026-09-19'],
    venue: 'Club Ciudad de Buenos Aires',
    capacity: 5000, available: 800, category: 'nacional',
    image_url: '/events/nacional4.png', price: 111000, vipPrice: 160000, status: 'active',
  },
  {
    ID: 9,
    title: 'Divididos',
    description: 'Ricardo Mollo y Diego Arnedo vuelven a los escenarios con el clásico sonido de Divididos.',
    date: '2026-07-04', dates: ['2026-07-04'],
    venue: 'Estadio Obras, Buenos Aires',
    capacity: 8000, available: 1500, category: 'nacional',
    image_url: '/events/nacional5.png', price: 55000, vipPrice: 100000, status: 'active',
  },
  {
    ID: 10,
    title: 'Maldita Felicidad',
    description: 'La comedia protagonizada por Pablo Echarri, Paola Krum, Carlos Portaluppi e Inés Palombo.',
    date: '2026-06-28', dates: ['2026-06-28'],
    venue: 'Teatro Ciudad de las Artes, Córdoba',
    capacity: 1200, available: 340, category: 'teatro',
    image_url: '/events/teatro1.png', price: 45000, status: 'active',
  },
  {
    ID: 11,
    title: 'Fito Páez — Segunda Gran Fiesta ¡FA!',
    description: 'El rosarino regresa al Movistar Arena para una segunda noche de su show ¡FA!',
    date: '2026-07-07', dates: ['2026-07-07'],
    venue: 'Movistar Arena, Buenos Aires',
    capacity: 15000, available: 2100, category: 'nacional',
    image_url: '/events/teatro2.png', price: 60000, vipPrice: 120000, status: 'active',
  },
  {
    ID: 12,
    title: 'Madama Butterfly',
    description: "La ópera de Puccini en una producción de Juventus Lyrica con dirección de Ana D'Anna.",
    date: '2026-07-11', dates: ['2026-07-11', '2026-07-12', '2026-07-18', '2026-07-19'],
    venue: 'Teatro Avenida, Buenos Aires',
    capacity: 900, available: 220, category: 'teatro',
    image_url: '/events/teatro3.png', price: 30000, status: 'active',
  },
  {
    ID: 13,
    title: 'Darío Orsi — Hasta las Manos',
    description: 'Stand up con Darío Orsi, el nuevo show del comediante argentino.',
    date: '2026-09-05', dates: ['2026-09-05'],
    venue: 'Teatro Gran Rex, Buenos Aires',
    capacity: 1200, available: 380, category: 'standup',
    image_url: '/events/standUp1.png', price: 20000, status: 'active',
  },
  {
    ID: 14,
    title: 'Manuel Ángel Redondo — Adulto Responsable',
    description: 'Stand up & Crowd Work en vivo en Córdoba.',
    date: '2026-08-30', dates: ['2026-08-30'],
    venue: 'Studio Theater, Córdoba',
    capacity: 500, available: 140, category: 'standup',
    image_url: '/events/standUp2.png', price: 20000, status: 'active',
  },
]

// Hero: exactamente estos 3 eventos
const HERO_IDS = [2, 3, 4]

// Primeros 3 eventos para el carousel del hero
const HERO_SLIDES = MOCK_EVENTS.slice(0, 3)

function HeroCarousel({ slides }) {
  const [current, setCurrent] = useState(0)
  const navigate = useNavigate()
  const timerRef = useRef(null)

  const start = () => {
    timerRef.current = setInterval(() => setCurrent(c => (c + 1) % slides.length), 4500)
  }

  useEffect(() => {
    start()
    return () => clearInterval(timerRef.current)
  }, [slides.length])

  const goTo = (i) => {
    clearInterval(timerRef.current)
    setCurrent(i)
    start()
  }

  const ev = slides[current]
  if (!ev) return null

  // Agrega T12:00:00 si es sólo fecha para evitar offset de zona horaria
  const parseDate = (d) => new Date(d.length === 10 ? d + 'T12:00:00' : d)
  const formatDate = (d) => parseDate(d).toLocaleDateString('es-AR', { weekday: 'long', day: '2-digit', month: 'long' })

  return (
    <div className={styles.hero}>
      {/* Imágenes con crossfade */}
      {slides.map((s, i) => (
        <div
          key={s.ID}
          className={`${styles.heroSlide} ${i === current ? styles.heroSlideActive : ''}`}
          style={{ backgroundImage: `url(${s.image_url})` }}
        />
      ))}
      <div className={styles.heroOverlay} />

      <div className={styles.heroContent}>
        <p className={styles.heroMeta}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
          {ev.venue}
        </p>
        <p className={styles.heroDate}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
          {formatDate(ev.date)}
        </p>
        <div className={styles.heroActions}>
          <button className={styles.heroBuyBtn} onClick={() => navigate(`/eventos/${ev.ID}`)}>
            Comprar entradas
          </button>
        </div>
      </div>

      {/* Dots */}
      <div className={styles.heroDots}>
        {slides.map((_, i) => (
          <button key={i} className={`${styles.heroDot} ${i === current ? styles.heroDotActive : ''}`} onClick={() => goTo(i)} />
        ))}
      </div>
    </div>
  )
}

export default function Home() {
  const [allEvents, setAllEvents] = useState([])
  const [loading, setLoading] = useState(true)
  const [searchText, setSearchText] = useState('')
  const [searchDate, setSearchDate] = useState('')
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()

  // La búsqueda vive en la URL → cambiar la URL limpia todo
  const qParam   = searchParams.get('q') || ''
  const dateParam = searchParams.get('date') || ''
  const catFilter = searchParams.get('cat') || ''
  const isSearching = qParam || dateParam

  // Sincronizar inputs con params de URL
  useEffect(() => { setSearchText(qParam) }, [qParam])
  useEffect(() => { setSearchDate(dateParam) }, [dateParam])

  useEffect(() => {
    fetchEvents()
  }, [])

  const fetchEvents = async () => {
    setLoading(true)
    try {
      const res = await getEvents({})
      const data = (res.data.data || []).map(e => ({
        ...e,
        dates: [e.date?.slice(0,10), ...(e.extra_dates ? e.extra_dates.split(',').map(d=>d.trim()).filter(Boolean) : [])].filter(Boolean).sort()
      }))
      setAllEvents(data.length > 0 ? data : MOCK_EVENTS)
    } catch {
      setAllEvents(MOCK_EVENTS)
    } finally {
      setLoading(false)
    }
  }

  // Normaliza texto quitando acentos y pasando a minúsculas
  const norm = (s) => (s || '').toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '')

  const handleSearchText = (e) => {
    e.preventDefault()
    if (!searchText.trim()) { navigate('/'); return }
    setSearchParams({ q: searchText.trim() })
  }

  const handleSearchDate = (e) => {
    e.preventDefault()
    if (!searchDate) { navigate('/'); return }
    setSearchParams({ date: searchDate })
  }

  const clearSearch = () => navigate('/')

  // Filtrado
  let displayed = [...allEvents]

  if (catFilter) {
    if (catFilter === 'musica') {
      displayed = displayed.filter(e => e.category === 'internacional' || e.category === 'nacional')
    } else {
      displayed = displayed.filter(e => e.category === catFilter)
    }
  }
  if (qParam) {
    const q = norm(qParam)
    // Etiquetas legibles por categoría — evita falsos positivos por venue o descripción
    const CAT_SEARCH = {
      internacional: 'musica internacional',
      nacional:      'musica nacional',
      teatro:        'teatro opera',
      standup:       'stand up comedia',
    }
    displayed = displayed.filter(e => {
      const catLabel = CAT_SEARCH[e.category] || ''
      return norm(e.title).includes(q) || catLabel.includes(q)
    })
  }
  if (dateParam) {
    displayed = displayed.filter(e => {
      const allDates = e.dates || [e.date]
      return allDates.some(d => d.slice(0, 10) === dateParam)
    })
  }

  const heroSlides = allEvents.filter(e => HERO_IDS.includes(e.ID))

  const CAT_LABELS = { musica: 'Música', internacional: 'Internacionales', nacional: 'Nacionales', teatro: 'Teatro y ópera', standup: 'Stand up' }

  if (loading) return (
    <div className={styles.loading}><div className={styles.spinner} /></div>
  )

  return (
    <div className={styles.page}>

      {/* ── Hero carousel — solo cuando no hay búsqueda ni filtro ── */}
      {!isSearching && !catFilter && <HeroCarousel slides={heroSlides} />}

      {/* ── Buscador ── */}
      <div className={styles.searchBar}>
        {/* Búsqueda por texto */}
        <form onSubmit={handleSearchText} className={styles.searchGroup}>
          <div className={styles.searchInputWrap}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#71717A" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            <input
              type="text"
              placeholder="Artista o evento..."
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              className={styles.searchInput}
            />
          </div>
          <button type="submit" className={styles.searchBtn}>Buscar</button>
        </form>

        {/* Búsqueda por fecha */}
        <form onSubmit={handleSearchDate} className={styles.searchGroup}>
          <div className={styles.searchInputWrap}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#71717A" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
            <input
              type="date"
              value={searchDate}
              onChange={e => setSearchDate(e.target.value)}
              className={`${styles.searchInput} ${styles.dateInput}`}
            />
          </div>
          <button type="submit" className={styles.searchBtn}>Buscar por fecha</button>
        </form>

        {isSearching && (
          <button type="button" onClick={clearSearch} className={styles.clearBtn}>
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Limpiar
          </button>
        )}
      </div>

      {/* ── Contenido ── */}
      <div className={styles.content}>

        {/* Encabezado de sección */}
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>
            {isSearching
              ? `Resultados (${displayed.length})`
              : catFilter
                ? CAT_LABELS[catFilter] || 'Eventos'
                : 'Todos los eventos'
            }
          </h2>
          {(isSearching || catFilter) && (
            <Link to="/" className={styles.backLink}>Ver todos</Link>
          )}
        </div>

        {/* Grid de eventos */}
        {displayed.length === 0 ? (
          <div className={styles.empty}>
            <p className={styles.emptyTitle}>No se encontraron eventos</p>
            <p className={styles.emptySub}>Probá con otro artista, fecha o categoría</p>
          </div>
        ) : (
          <div className={styles.grid}>
            {displayed.map(ev => <EventCard key={ev.ID} event={ev} grid />)}
          </div>
        )}
      </div>
    </div>
  )
}
