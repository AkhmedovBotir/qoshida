export type IdentificationStatus = 'none' | 'pending' | 'approved' | 'rejected'

export type IdentDocumentKind = 'diploma' | 'certificate' | 'license' | 'other'

export type IdentDocument = {
  id: string
  kind: IdentDocumentKind
  title: string
  path: string
  file_name: string
}

export type Identification = {
  id?: string
  provider_id: string
  full_name?: string
  pinfl?: string
  passport?: string
  birth_date?: string
  address?: string
  experience_years?: number
  about?: string
  passport_image?: string
  selfie_image?: string
  documents?: IdentDocument[]
  status: IdentificationStatus
  rejection_note?: string
  reviewed_by_name?: string
  reviewed_by_role?: string
  reviewed_at?: string
  submitted_at?: string
  provider_name?: string
  phone?: string
  activity_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
  created_at?: string
  audit?: import('../lib/audit').AuditTrail
}

export type IdentificationListResult = {
  items: Identification[]
  total: number
  page: number
  limit: number
}

export type IdentDocumentInput = {
  kind: IdentDocumentKind
  title: string
  file_name: string
  path?: string
  content?: string
}

export type IdentificationSubmit = {
  full_name: string
  pinfl: string
  passport: string
  birth_date: string
  address: string
  experience_years: number
  about: string
  passport_image: string
  selfie_image: string
  documents: IdentDocumentInput[]
}

export const identStatusLabel: Record<IdentificationStatus, string> = {
  none: 'Topshirilmagan',
  pending: 'Kutilmoqda',
  approved: 'Tasdiqlangan',
  rejected: 'Bekor qilingan',
}

export const identDocKindLabel: Record<IdentDocumentKind, string> = {
  diploma: 'Diplom',
  certificate: 'Sertifikat',
  license: 'Litsenziya',
  other: 'Boshqa',
}
