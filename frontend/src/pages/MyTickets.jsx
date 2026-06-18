import { useState, useEffect } from 'react'
import { getMyTickets, cancelTicket, transferTicket } from '../services/api'
import { useAuth } from '../context/AuthContext'
import styles from './MyTickets.module.css'

function groupTicketsByEvent(tickets) {
  // Cancelled tickets are hidden entirely
  const visible = tickets.filter(t => t.status !== 'cancelled')
  const groups = {}
  for (const t of visible) {
    const key = t.EventID || t.event?.ID || t.event_id || 'unknown'
    if (!groups[key]) {
      groups[key] = { event: t.event, tickets: [] }
    }
    groups[key].tickets.push(t)
  }
  // Remove groups with no tickets left
  return Object.values(groups).filter(g => g.tickets.length > 0)
}

export default function MyTickets() {
  const { user } = useAuth()
  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [transferModal, setTransferModal] = useState(null) // ticketId
  const [transferEmail, setTransferEmail] = useState('')
  const [transferError, setTransferError] = useState('')
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

  useEffect(() => { fetchTickets() }, [])

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
    setTransferError('')
    try {
      await transferTicket(transferModal, transferEmail)
      setMessage({ type: 'success', text: 'Entrada traspasada exitosamente. Se notificó al nuevo titular por email.' })
      setTransferModal(null)
      setTransferEmail('')
      fetchTickets()
    } catch (err) {
      setTransferError(err.response?.data?.error || 'Error al traspasar')
    } finally {
      setActionLoading(false)
    }
  }

  if (loading) return <div className={styles.loading}>Cargando tus entradas...</div>

  const groups = groupTicketsByEvent(tickets)
  const totalVisible = tickets.filter(t => t.status === 'active').length
  const statusLabel = { active: 'Validada', transferred: 'Traspasada' }
  const statusClass = { active: styles.statusActive, transferred: styles.statusTransferred }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1>Mis Entradas</h1>
        <p className={styles.subtitle}>
          {totalVisible} entrada{totalVisible !== 1 ? 's' : ''} encontrada{totalVisible !== 1 ? 's' : ''}
        </p>
      </div>

      {message && (
        <div className={`${styles.message} ${message.type === 'success' ? styles.msgSuccess : styles.msgError}`}>
          {message.text}
          <button onClick={() => setMessage(null)}>✕</button>
        </div>
      )}

      {groups.length === 0 ? (
        <div className={styles.empty}>
          <p>No tenés entradas aún</p>
        </div>
      ) : (
        <div className={styles.ticketList}>
          {groups.map((group, gi) => {
            const { event, tickets: groupTickets } = group
            const count = groupTickets.length
            const activeTickets = groupTickets.filter(t => t.status === 'active')
            const priceEach = groupTickets[0]?.price || event?.price || 0
            const totalAmount = priceEach * activeTickets.length
            const totalStr = totalAmount === 0 ? 'Gratis' : `$${Number(totalAmount).toLocaleString('es-AR')}`
            const priceEachStr = priceEach === 0 ? 'Gratis' : `$${Number(priceEach).toLocaleString('es-AR')} c/u`
            const ticketType = (groupTickets[0]?.ticket_type || 'general').toUpperCase()
            const allActive = groupTickets.every(t => t.status === 'active')

            return (
              <div key={gi} className={styles.ticketCard}>
                <div className={styles.ticketStrip} />

                <div className={styles.ticketBody}>
                  <div className={styles.ticketMain}>
                    <div className={styles.eventTitle}>{event?.title || 'Evento'}</div>

                    <div className={styles.statusRow}>
                      {groupTickets.map(t => (
                        <span key={t.ID} className={`${styles.status} ${statusClass[t.status] || styles.statusActive}`}>
                          {statusLabel[t.status] || t.status}
                        </span>
                      ))}
                    </div>

                    <div className={styles.buyerInfo}>
                      {user?.name} — {user?.email}
                    </div>

                    <div className={styles.ticketType}>{ticketType}</div>

                    <div className={styles.priceBlock}>
                      <span className={styles.totalAmount}>{totalStr}</span>
                      {count > 1 && (
                        <span className={styles.countBadge}>{count} entradas · {priceEachStr}</span>
                      )}
                    </div>

                    <div className={styles.metaRow}>
                      <span>{event?.venue}</span>
                    </div>

                    <div className={styles.idRow}>
                      <div className={styles.idBlock}>
                        <span className={styles.idLabel}>ID Compra</span>
                        <span className={styles.idValue}>{String(groupTickets[0]?.ID).padStart(9, '0')}</span>
                      </div>
                    </div>
                  </div>

                  {/* QR section — one per ticket */}
                  <div className={styles.qrSection}>
                    <div className={styles.dashedDivider} />
                    {groupTickets.map((ticket) => (
                      <div key={ticket.ID} className={styles.qrBlock}>
                        {ticket.status === 'active' && ticket.qr_code ? (
                          <img src={ticket.qr_code} alt="QR Code" className={styles.qr} />
                        ) : (
                          <div className={styles.qrPlaceholder}>
                            {ticket.status === 'transferred' ? 'Traspasada' : ticket.status === 'cancelled' ? 'Cancelada' : 'Sin QR'}
                          </div>
                        )}
                        <span className={styles.qrLabel}>Código QR</span>
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
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
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
            {transferError && (
              <div className={styles.modalError}>{transferError}</div>
            )}
            <div className={styles.modalActions}>
              <button onClick={handleTransfer} disabled={actionLoading || !transferEmail} className={styles.confirmBtn}>
                {actionLoading ? 'Traspasando...' : 'Confirmar traspaso'}
              </button>
              <button onClick={() => { setTransferModal(null); setTransferEmail(''); setTransferError('') }} className={styles.cancelModalBtn}>
                Cancelar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
