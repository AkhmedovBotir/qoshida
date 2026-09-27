import { api } from './api'
import type { Identification, IdentificationSubmit } from '../types/identification'

export const IDENT_BASE = '/api/v1/service-provider/identification'

export function getOwnIdentification() {
  return api<Identification>(IDENT_BASE)
}

export function submitIdentification(body: IdentificationSubmit) {
  return api<Identification>(IDENT_BASE, { method: 'PUT', body, timeoutMs: 60_000 })
}
