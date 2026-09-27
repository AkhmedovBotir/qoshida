export type AuditEntry = {
  name: string
  role: string
  at: string
}

export type AuditTrail = {
  created?: AuditEntry
  approved?: AuditEntry
  rejected?: AuditEntry
  updates?: AuditEntry[]
}

function formatAt(at?: string) {
  if (!at) return ''
  const date = new Date(at)
  if (Number.isNaN(date.getTime())) return at
  return date.toLocaleString('uz-UZ')
}

export function auditFields(audit?: AuditTrail | null) {
  const fields: { label: string; value: string }[] = []
  if (audit?.created) {
    fields.push({
      label: 'Yaratgan',
      value: `${audit.created.role} — ${audit.created.name}${audit.created.at ? ` (${formatAt(audit.created.at)})` : ''}`,
    })
  }
  if (audit?.approved) {
    fields.push({
      label: 'Tasdiqlagan',
      value: `${audit.approved.role} — ${audit.approved.name}${audit.approved.at ? ` (${formatAt(audit.approved.at)})` : ''}`,
    })
  }
  if (audit?.rejected) {
    fields.push({
      label: 'Bekor qilgan',
      value: `${audit.rejected.role} — ${audit.rejected.name}${audit.rejected.at ? ` (${formatAt(audit.rejected.at)})` : ''}`,
    })
  }
  if (audit?.updates?.length) {
    fields.push({
      label: 'O‘zgartirganlar',
      value: audit.updates
        .map((item) => `${item.role} — ${item.name}${item.at ? ` (${formatAt(item.at)})` : ''}`)
        .join('\n'),
    })
  }
  return fields
}
