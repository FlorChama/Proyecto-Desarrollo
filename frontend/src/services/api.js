import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Auth
export const register = (data) => api.post('/auth/register', data)
export const login = (data) => api.post('/auth/login', data)

// Events
export const getEvents = (params) => api.get('/events', { params })
export const getEventById = (id) => api.get(`/events/${id}`)

// Tickets
export const buyTicket = (eventId, paymentMethod, amount, ticketType = 'general') =>
  api.post('/tickets', { event_id: eventId, payment_method: paymentMethod, amount, ticket_type: ticketType })
export const getMyTickets = () => api.get('/tickets/my')
export const getMyPayments = () => api.get('/tickets/payments')
export const cancelTicket = (id) => api.delete(`/tickets/${id}`)
export const transferTicket = (id, targetEmail) => api.post(`/tickets/${id}/transfer`, { target_email: targetEmail })

// Admin
export const getAdminEvents = () => api.get('/admin/events')
export const uploadEventImage = (file) => {
  const fd = new FormData()
  fd.append('image', file)
  return api.post('/admin/events/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
}
export const createEvent = (data) => api.post('/admin/events', data)
export const updateEvent = (id, data) => api.put(`/admin/events/${id}`, data)
export const deleteEvent = (id) => api.delete(`/admin/events/${id}`)
export const getEventReport = (id) => api.get(`/admin/events/${id}/report`)

export default api
