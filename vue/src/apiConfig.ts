export const API_BASE_URL = 'http://localhost:3000';

export function authHeaders() {
  const token = localStorage.getItem('jwt')
  return token ? { Authorization: `Bearer ${token}` } : {}
} 