import { useState, useEffect } from 'react'
import { getMyTickets, cancelTicket, transferTicket } from '../services/api'
import styles from './MyTickets.module.css'

export default function MyTickets() {
  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [transferModal, setTransferModal] = useState(null) // ticketId
  const [transferEmail, setTransferEmail] = useState('')
  const [actionLoading, setActionLoading] = useState(false)
  const [message, setMessage] = useState(null)

  const fetchTickets = async () => {
    setLoading(true)
    try {
      const res = await getMyTickets()
      setTickets(res.data.data || [])
    } catch {
      setTickets([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTickets()
  }, [])

  const handleCancel = async (ticketId) => {
    if (!confirm('¿Estás seguro de que querés cancelar esta entrada?')) return
    setActionLoading(true)
    try {
      await cancelTicket(ticketId)
      setMessage({ type: 'success', text: 'Entrada cancelada exitosamente' })
      fetchTickets()
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || 'Error al cancelar' })
    } finally {
      setActionLoading(false)
    }
  }

  const handleTransfer = async () => {
    if (!transferEmail) return
    setActionLoading(true)
    try {
      await transferTicket(transferModal, transferEmail)
      setMessage({ type: 'success', text: 'Entrada traspasada exitosamente. Se notificó al nuevo titular por email.' })
      setTransferModal(null)
      setTransferEmail('')
      fetchTickets()
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || 'Error al traspasar' })
    } finally {
      setActionLoading(false)
    }
  }

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('es-AR', { day: '2-digit', month: 'long', year: 'numeric' })
  }

  const statusLabel = { active: 'Activa', cancelled: 'Cancelada', transferred: 'Traspasada' }
  const statusClass = { active: styles.statusActive, cancelled: styles.statusCancelled, transferred: styles.statusTransferred }

  if (loading) return <div className={styles.loading}>Cargando tus entradas...</div>

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1>Mis Entradas</h1>
        <p className={styles.subtitle}>{tickets.length} entrada{tickets.length !== 1 ? 's' : ''} encontrada{tickets.length !== 1 ? 's' : ''}</p>
      </div>

      {message && (
        <div className={`${styles.message} ${message.type === 'success' ? styles.msgSuccess : styles.msgError}`}>
          {message.text}
          <button onClick={() => setMessage(null)}>✕</button>
        </div>
      )}

      {tickets.length === 0 ? (
        <div className={styles.empty}>
          <span>🎫</span>
          <p>No tenés entradas aún</p>
        </div>
      ) : (
        <div className={styles.ticketList}>
          {tickets.map((ticket) => (
            <div key={ticket.ID} className={styles.ticketCard}>
              <div className={styles.ticketLeft}>
                <div className={styles.ticketEvent}>{ticket.event?.title || 'Evento'}</div>
                <div className={styles.ticketDetails}>
                  <span>📍 {ticket.event?.venue}</span>
                  <span>📅 {ticket.event?.date ? formatDate(ticket.event.date) : ''}</span>
                  <span>💲 {ticket.event?.price === 0 ? 'Gratis' : `$${ticket.event?.price?.toLocaleString('es-AR')}`}</span>
                </div>
                <span className={`${styles.status} ${statusClass[ticket.status]}`}>
                  {statusLabel[ticket.status]}
                </span>
              </div>

              <div className={styles.ticketRight}>
                {ticket.qr_code && (
                  <img src={ticket.qr_code} alt="QR" className={styles.qr} />
                )}
                {ticket.status === 'active' && (
                  <div className={styles.actions}>
                    <button
                      className={styles.transferBtn}
                      onClick={() => { setTransferModal(ticket.ID); setMessage(null) }}
                      disabled={actionLoading}
                    >
                      Traspasar
                    </button>
                    <button
                      className={styles.cancelBtn}
                      onClick={() => handleCancel(ticket.ID)}
                      disabled={actionLoading}
                    >
                      Cancelar
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {transferModal && (
        <div className={styles.modalOverlay} onClick={(e) => e.target === e.currentTarget && setTransferModal(null)}>
          <div className={styles.modal}>
            <h2>Traspasar entrada</h2>
            <p>Ingresá el email del usuario al que querés transferirle la entrada.</p>
            <input
              type="email"
              placeholder="email@ejemplo.com"
              value={transferEmail}
              onChange={(e) => setTransferEmail(e.target.value)}
              className={styles.modalInput}
            />
            <div className={styles.modalActions}>
              <button onClick={handleTransfer} disabled={actionLoading || !transferEmail} className={styles.confirmBtn}>
                {actionLoading ? 'Traspasando...' : 'Confirmar traspaso'}
              </button>
              <button onClick={() => { setTransferModal(null); setTransferEmail('') }} className={styles.cancelModalBtn}>
                Cancelar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
