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
  const [onlyAvailable, setOnlyAvailable] = useState(false)

  const fetchEvents = async () => {
    setLoading(true)
    try {
      const params = {}
      if (search) params.search = search
      if (category && category !== 'Todos') params.category = category
      if (onlyAvailable) params.available = true

      const res = await getEvents(params)
      setEvents(res.data.data || [])
    } catch {
      setEvents([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchEvents()
  }, [category, onlyAvailable])

  const handleSearch = (e) => {
    e.preventDefault()
    fetchEvents()
  }

  return (
    <div className={styles.page}>
      <div className={styles.hero}>
        <h1 className={styles.heroTitle}>Encontrá tu próximo evento</h1>
        <p className={styles.heroSubtitle}>Comprá entradas para los mejores shows, festivales y más</p>

        <form onSubmit={handleSearch} className={styles.searchForm}>
          <input
            type="text"
            placeholder="Buscar eventos, artistas, venues..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className={styles.searchInput}
          />
          <button type="submit" className={styles.searchBtn}>Buscar</button>
        </form>
      </div>

      <div className={styles.filters}>
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

        <label className={styles.availableToggle}>
          <input
            type="checkbox"
            checked={onlyAvailable}
            onChange={(e) => setOnlyAvailable(e.target.checked)}
          />
          Solo con disponibilidad
        </label>
      </div>

      <div className={styles.content}>
        {loading ? (
          <div className={styles.loading}>Cargando eventos...</div>
        ) : events.length === 0 ? (
          <div className={styles.empty}>
            <span>🎭</span>
            <p>No se encontraron eventos</p>
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
