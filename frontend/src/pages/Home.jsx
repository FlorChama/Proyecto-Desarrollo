import { useState, useEffect } from 'react'
import { getEvents } from '../services/api'
import EventCard from '../components/EventCard'
import styles from './Home.module.css'

const CATEGORIES = ['Todos', 'Música', 'Teatro', 'Deporte', 'Arte', 'Tecnología', 'Otro']

export default function Home() {
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')
  const [location, setLocation] = useState('')
  const [date, setDate] = useState('')
  const [onlyAvailable, setOnlyAvailable] = useState(false)

  const fetchEvents = async () => {
    setLoading(true)
    try {
      const params = {}
      if (search) params.search = search
      if (category && category !== 'Todos') params.category = category
      if (location) params.search = location
      if (onlyAvailable) params.available = true
      const res = await getEvents(params)
      let data = res.data.data || []
      if (date) {
        data = data.filter(e => new Date(e.date).toISOString().slice(0, 10) === date)
      }
      setEvents(data)
    } catch {
      setEvents([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchEvents() }, [category, onlyAvailable])

  const handleSearch = (e) => {
    e.preventDefault()
    fetchEvents()
  }

  const clearFilters = () => {
    setSearch(''); setCategory(''); setLocation(''); setDate(''); setOnlyAvailable(false)
    setTimeout(fetchEvents, 0)
  }

  const hasFilters = search || (category && category !== 'Todos') || location || date || onlyAvailable

  return (
    <div className={styles.page}>

      {/* Hero */}
      <div className={styles.hero}>
        <div className={styles.heroBg} />
        <div className={styles.heroContent}>
          <p className={styles.heroEyebrow}>🎫 La mejor experiencia en vivo</p>
          <h1 className={styles.heroTitle}>Encontrá tu<br /><span className={styles.heroAccent}>próximo show</span></h1>
          <p className={styles.heroSub}>Entradas para los mejores conciertos, festivales y eventos de Argentina</p>

          <form onSubmit={handleSearch} className={styles.searchBox}>
            <div className={styles.searchField}>
              <span className={styles.searchIcon}>🔍</span>
              <input
                type="text"
                placeholder="Artista, evento o festival..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className={styles.searchInput}
              />
            </div>
            <div className={styles.searchDivider} />
            <div className={styles.searchField}>
              <span className={styles.searchIcon}>📍</span>
              <input
                type="text"
                placeholder="Ciudad o venue..."
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                className={styles.searchInput}
              />
            </div>
            <div className={styles.searchDivider} />
            <div className={styles.searchField}>
              <span className={styles.searchIcon}>📅</span>
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className={`${styles.searchInput} ${styles.dateInput}`}
              />
            </div>
            <button type="submit" className={styles.searchBtn}>Buscar</button>
          </form>
        </div>
      </div>

      {/* Filtros */}
      <div className={styles.filtersBar}>
        <div className={styles.categories}>
          {CATEGORIES.map((cat) => (
            <button
              key={cat}
              className={`${styles.catBtn} ${category === (cat === 'Todos' ? '' : cat) ? styles.catActive : ''}`}
              onClick={() => setCategory(cat === 'Todos' ? '' : cat)}
            >
              {cat}
            </button>
          ))}
        </div>
        <div className={styles.filterRight}>
          <label className={styles.availableToggle}>
            <input type="checkbox" checked={onlyAvailable} onChange={(e) => setOnlyAvailable(e.target.checked)} />
            Solo disponibles
          </label>
          {hasFilters && (
            <button onClick={clearFilters} className={styles.clearBtn}>✕ Limpiar</button>
          )}
        </div>
      </div>

      {/* Contenido */}
      <div className={styles.content}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>
            {hasFilters ? 'Resultados' : 'Próximos shows'}
          </h2>
          {!loading && <span className={styles.eventCount}>{events.length} eventos</span>}
        </div>

        {loading ? (
          <div className={styles.loading}>
            <div className={styles.spinner} />
            <p>Cargando eventos...</p>
          </div>
        ) : events.length === 0 ? (
          <div className={styles.empty}>
            <span>🎭</span>
            <p>No se encontraron eventos</p>
            {hasFilters && <button onClick={clearFilters} className={styles.emptyBtn}>Ver todos los eventos</button>}
          </div>
        ) : (
          <div className={styles.grid}>
            {events.map((event) => (
              <EventCard key={event.ID} event={event} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
