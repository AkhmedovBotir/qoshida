import { useEffect, useState } from 'react'
import { Select } from '../ui/Select'
import { listPublicRegions } from '../../lib/market'
import type { RegionItem } from '../../types/market'

type GeoValue = {
  region_id: string
  district_id: string
  mfy_id: string
}

export function GeoFields({
  value,
  onChange,
  required = false,
}: {
  value: GeoValue
  onChange: (value: GeoValue) => void
  required?: boolean
}) {
  const [regions, setRegions] = useState<RegionItem[]>([])
  const [districts, setDistricts] = useState<RegionItem[]>([])
  const [mfys, setMfys] = useState<RegionItem[]>([])

  useEffect(() => {
    void listPublicRegions({ type: 'region', limit: 50 }).then((data) => setRegions(data.items ?? [])).catch(() => setRegions([]))
  }, [])

  useEffect(() => {
    if (!value.region_id) {
      setDistricts([])
      return
    }
    void listPublicRegions({ type: 'district', parent_id: value.region_id, limit: 200 })
      .then((data) => setDistricts(data.items ?? []))
      .catch(() => setDistricts([]))
  }, [value.region_id])

  useEffect(() => {
    if (!value.district_id) {
      setMfys([])
      return
    }
    void listPublicRegions({ type: 'mfy', parent_id: value.district_id, limit: 300 })
      .then((data) => setMfys(data.items ?? []))
      .catch(() => setMfys([]))
  }, [value.district_id])

  return (
    <div className="space-y-3">
      <Select
        label="Viloyat"
        value={value.region_id}
        onChange={(region_id) => onChange({ region_id, district_id: '', mfy_id: '' })}
        options={regions.map((item) => ({ value: item.id, label: item.name }))}
        placeholder="Viloyatni tanlang"
        required={required}
      />
      <Select
        label="Tuman / shahar"
        value={value.district_id}
        onChange={(district_id) => onChange({ ...value, district_id, mfy_id: '' })}
        options={districts.map((item) => ({ value: item.id, label: item.name }))}
        placeholder="Tumanni tanlang"
        disabled={!value.region_id}
        required={required}
      />
      <Select
        label="MFY"
        value={value.mfy_id}
        onChange={(mfy_id) => onChange({ ...value, mfy_id })}
        options={mfys.map((item) => ({ value: item.id, label: item.name }))}
        placeholder="MFY ni tanlang"
        disabled={!value.district_id}
        required={required}
      />
    </div>
  )
}
