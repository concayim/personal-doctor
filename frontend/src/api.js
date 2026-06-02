const apiBase = import.meta.env.VITE_API_BASE || ''
const jsonHeaders = { 'Content-Type': 'application/json' }

async function request(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, options)
  const data = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(data?.error || `请求失败：${response.status}`)
  }
  return data
}

export function listPatients() {
  return request('/api/patients').then(asArray)
}

export function createPatient(patient) {
  return request('/api/patients', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify(patient),
  })
}

export function listRecords(patientId) {
  return request(`/api/patients/${patientId}/records`).then(asArray)
}

export function createRecord(patientId, record) {
  return request(`/api/patients/${patientId}/records`, {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify(record),
  })
}

export function listMessages(patientId) {
  return request(`/api/patients/${patientId}/messages`).then(asArray)
}

export function sendMessage(patientId, message) {
  return request(`/api/patients/${patientId}/chat`, {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ message }),
  })
}

function asArray(value) {
  return Array.isArray(value) ? value : []
}
