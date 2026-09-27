import { FileText, Loader2, Plus, Trash2 } from 'lucide-react'
import { toast } from '../lib/snack'
import { FormEvent, useEffect, useState, type ReactNode } from 'react'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { Select } from '../components/ui/Select'
import { APP_COLOR, API_URL } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { api, ApiError } from '../lib/api'
import { getOwnIdentification, submitIdentification } from '../lib/identifications'
import type { ProviderUser } from '../types/provider'
import {
  identDocKindLabel,
  identStatusLabel,
  type IdentDocumentInput,
  type IdentDocumentKind,
  type Identification,
  type IdentificationStatus,
} from '../types/identification'

type DocRow = IdentDocumentInput & { key: string }

const emptyDoc = (): DocRow => ({
  key: crypto.randomUUID(),
  kind: 'diploma',
  title: '',
  file_name: '',
  path: '',
  content: '',
})

function readFile(file: File, maxBytes: number, kind: 'image' | 'pdf'): Promise<string> {
  return new Promise((resolve, reject) => {
    if (file.size > maxBytes) {
      reject(new Error(kind === 'pdf' ? 'PDF 5 MB dan oshmasin' : 'Rasm 10 MB dan oshmasin'))
      return
    }
    if (kind === 'pdf' && file.type !== 'application/pdf') {
      reject(new Error('Hujjat faqat PDF bo‘lishi kerak'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(new Error('Faylni o‘qib bo‘lmadi'))
    reader.readAsDataURL(file)
  })
}

function fileUrl(path?: string) {
  if (!path) return ''
  if (path.startsWith('data:')) return path
  return `${API_URL}${path}`
}

function StatusBadge({ status }: { status: IdentificationStatus }) {
  const cls =
    status === 'approved'
      ? 'bg-emerald-50 text-emerald-700'
      : status === 'rejected'
        ? 'bg-red-50 text-red-700'
        : status === 'pending'
          ? 'bg-amber-50 text-amber-700'
          : 'bg-slate-100 text-slate-600'
  return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${cls}`}>{identStatusLabel[status]}</span>
}

export function IdentificationPage() {
  const { setUser } = useAuth()
  const [item, setItem] = useState<Identification | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const [fullName, setFullName] = useState('')
  const [pinfl, setPinfl] = useState('')
  const [passport, setPassport] = useState('')
  const [birthDate, setBirthDate] = useState('')
  const [address, setAddress] = useState('')
  const [experience, setExperience] = useState('0')
  const [about, setAbout] = useState('')
  const [passportImage, setPassportImage] = useState('')
  const [selfieImage, setSelfieImage] = useState('')
  const [docs, setDocs] = useState<DocRow[]>([emptyDoc()])

  const status = item?.status ?? 'none'
  const locked = status === 'pending' || status === 'approved'
  const canEdit = status === 'none' || status === 'rejected'

  useEffect(() => {
    let cancelled = false
    getOwnIdentification()
      .then((data) => {
        if (cancelled) return
        setItem(data)
        fillForm(data)
      })
      .catch((err: unknown) => {
        if (!cancelled) if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  function fillForm(data: Identification) {
    setFullName(data.full_name ?? '')
    setPinfl(data.pinfl ?? '')
    setPassport(data.passport ?? '')
    setBirthDate(data.birth_date ?? '')
    setAddress(data.address ?? '')
    setExperience(String(data.experience_years ?? 0))
    setAbout(data.about ?? '')
    setPassportImage(data.passport_image ?? '')
    setSelfieImage(data.selfie_image ?? '')
    setDocs(
      data.documents?.length
        ? data.documents.map((doc) => ({
            key: doc.id || crypto.randomUUID(),
            kind: doc.kind,
            title: doc.title,
            file_name: doc.file_name,
            path: doc.path,
            content: '',
          }))
        : [emptyDoc()],
    )
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    if (!canEdit) return

    setSaving(true)
    try {
      const saved = await submitIdentification({
        full_name: fullName,
        pinfl,
        passport,
        birth_date: birthDate,
        address,
        experience_years: Number(experience) || 0,
        about,
        passport_image: passportImage,
        selfie_image: selfieImage,
        documents: docs.map(({ kind, title, file_name, path, content }) => ({
          kind,
          title,
          file_name,
          path,
          content,
        })),
      })
      setItem(saved)
      fillForm(saved)
      const me = await api<ProviderUser>('/api/v1/service-provider-auth/me')
      setUser(me)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
        <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <div className="flex flex-wrap items-center gap-3">
          <h2 className="text-2xl font-bold text-slate-900">Identifikatsiya</h2>
          <StatusBadge status={status} />
        </div>
        <p className="mt-1 text-sm text-slate-500">
          {status === 'approved'
            ? 'Ma’lumotlaringiz tasdiqlangan. Endi xizmat joylashingiz mumkin.'
            : status === 'pending'
              ? 'Ariza yuborildi. Savdo uyi rahbari yoki menejer tasdiqlashini kuting.'
              : 'Xizmat joylashdan oldin shaxsingiz va diplom/sertifikatlaringizni yuboring.'}
        </p>
      </div>

      {status === 'rejected' && item?.rejection_note ? (
        <div className="rounded-xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
          Bekor qilish sababi: {item.rejection_note}
        </div>
      ) : null}


      <form onSubmit={(e) => void onSubmit(e)} className="space-y-5 rounded-2xl border border-slate-200 bg-white p-5">
        <Field label="F.I.Sh">
          <input required disabled={locked} value={fullName} onChange={(e) => setFullName(e.target.value)} className={inputCls} placeholder="Aliyev Ali Aliyevich" />
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="JSHSHIR">
            <input
              required
              disabled={locked}
              inputMode="numeric"
              maxLength={14}
              value={pinfl}
              onChange={(e) => setPinfl(e.target.value.replace(/\D/g, '').slice(0, 14))}
              className={inputCls}
              placeholder="14 ta raqam"
            />
          </Field>
          <Field label="Pasport seria va raqami">
            <input
              required
              disabled={locked}
              maxLength={9}
              value={passport}
              onChange={(e) => setPassport(e.target.value.replace(/[^a-zA-Z0-9]/g, '').toUpperCase().slice(0, 9))}
              className={inputCls}
              placeholder="AA1234567"
            />
          </Field>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Tug‘ilgan sana">
            <input required disabled={locked} type="date" value={birthDate} onChange={(e) => setBirthDate(e.target.value)} className={inputCls} />
          </Field>
          <Field label="Ish tajribasi (yil)">
            <input
              required
              disabled={locked}
              type="number"
              min={0}
              max={70}
              value={experience}
              onChange={(e) => setExperience(e.target.value)}
              className={inputCls}
            />
          </Field>
        </div>
        <Field label="Yashash manzili">
          <input required disabled={locked} value={address} onChange={(e) => setAddress(e.target.value)} className={inputCls} placeholder="Ko‘cha, uy, xonadon" />
        </Field>
        <Field label="O‘zingiz haqingizda (ixtiyoriy)">
          <textarea disabled={locked} rows={3} value={about} onChange={(e) => setAbout(e.target.value)} className={inputCls} placeholder="Mutaxassislik, tajriba" />
        </Field>

        <div className="grid gap-4 sm:grid-cols-2">
          <PhotoField
            label="Pasport rasmi"
            value={passportImage}
            locked={locked}
            onChange={setPassportImage}
          />
          <PhotoField label="Selfi" value={selfieImage} locked={locked} onChange={setSelfieImage} />
        </div>

        <div>
          <div className="mb-3 flex items-center justify-between">
            <p className="text-sm font-semibold text-slate-800">Diplom va sertifikatlar (PDF)</p>
            {canEdit ? (
              <button
                type="button"
                onClick={() => setDocs((rows) => (rows.length >= 5 ? rows : [...rows, emptyDoc()]))}
                className="inline-flex items-center gap-1 text-sm font-medium text-sky-800"
              >
                <Plus size={14} /> Qo‘shish
              </button>
            ) : null}
          </div>
          <div className="space-y-3">
            {docs.map((doc, index) => (
              <div key={doc.key} className="rounded-xl border border-slate-200 p-3">
                <div className="grid gap-3 sm:grid-cols-[9rem_1fr_auto]">
                  <Select
                    value={doc.kind}
                    disabled={locked}
                    onChange={(v) =>
                      setDocs((rows) => rows.map((row) => (row.key === doc.key ? { ...row, kind: v as IdentDocumentKind } : row)))
                    }
                    options={[
                      { value: 'diploma', label: identDocKindLabel.diploma },
                      { value: 'certificate', label: identDocKindLabel.certificate },
                      { value: 'license', label: identDocKindLabel.license },
                      { value: 'other', label: identDocKindLabel.other },
                    ]}
                  />
                  <input
                    required
                    disabled={locked}
                    value={doc.title}
                    onChange={(e) =>
                      setDocs((rows) => rows.map((row) => (row.key === doc.key ? { ...row, title: e.target.value } : row)))
                    }
                    className={inputCls}
                    placeholder="Hujjat nomi"
                  />
                  {canEdit && docs.length > 1 ? (
                    <button
                      type="button"
                      onClick={() => setDocs((rows) => rows.filter((row) => row.key !== doc.key))}
                      className="justify-self-end rounded-lg p-2 text-red-600 hover:bg-red-50"
                    >
                      <Trash2 size={16} />
                    </button>
                  ) : (
                    <span />
                  )}
                </div>
                <div className="mt-3 flex flex-wrap items-center gap-3">
                  {canEdit ? (
                    <label className="inline-flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm">
                      <FileText size={16} />
                      PDF tanlash
                      <input
                        type="file"
                        accept="application/pdf"
                        className="hidden"
                        onChange={(e) => {
                          const file = e.target.files?.[0]
                          e.target.value = ''
                          if (!file) return
                          void readFile(file, 5 * 1024 * 1024, 'pdf')
                            .then((content) =>
                              setDocs((rows) =>
                                rows.map((row) =>
                                  row.key === doc.key ? { ...row, content, file_name: file.name, path: '' } : row,
                                ),
                              ),
                            )
                            .catch((err: unknown) => toast.error(err instanceof Error ? err.message : 'Fayl xato'))
                        }}
                      />
                    </label>
                  ) : null}
                  {doc.file_name || doc.path ? (
                    <a
                      href={doc.content || fileUrl(doc.path)}
                      target="_blank"
                      rel="noreferrer"
                      className="text-sm font-medium text-sky-800"
                    >
                      {doc.file_name || `Hujjat ${index + 1}`}
                    </a>
                  ) : (
                    <span className="text-sm text-slate-400">PDF tanlanmagan</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {canEdit ? (
          <button
            type="submit"
            disabled={saving}
            className="rounded-xl px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60"
            style={{ backgroundColor: APP_COLOR }}
          >
            {saving ? 'Yuborilmoqda...' : 'Tekshiruvga yuborish'}
          </button>
        ) : null}
      </form>
    </div>
  )
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-sm font-medium text-slate-700">{label}</span>
      {children}
    </label>
  )
}

function PhotoField({
  label,
  value,
  locked,
  onChange,
}: {
  label: string
  value: string
  locked: boolean
  onChange: (v: string) => void
}) {
  return (
    <div>
      <p className="mb-1.5 text-sm font-medium text-slate-700">{label}</p>
      {value ? <ImageThumb src={fileUrl(value)} className="mb-2 h-24 w-24 rounded-xl object-cover" /> : null}
      {locked ? null : (
        <label className="inline-flex cursor-pointer rounded-xl border border-slate-200 px-3 py-2 text-sm">
          Rasm tanlash
          <input
            type="file"
            accept="image/png,image/jpeg,image/webp"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0]
              e.target.value = ''
              if (!file) return
              void readFile(file, 10 * 1024 * 1024, 'image')
                .then(onChange)
                .catch((err: unknown) => {
                  toast.error(err instanceof Error ? err.message : 'Rasm xato')
                })
            }}
          />
        </label>
      )}
    </div>
  )
}

const inputCls =
  'w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-slate-400 disabled:bg-slate-50'
