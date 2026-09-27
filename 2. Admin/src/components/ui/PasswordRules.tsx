type PasswordRulesProps = {
  value: string
}

export function checkPasswordRules(value: string) {
  const length = [...value].length
  return {
    length: length >= 10 && length <= 72,
    letter: /\p{L}/u.test(value),
    digit: /\p{N}/u.test(value),
  }
}

export function isPasswordValid(value: string) {
  const rules = checkPasswordRules(value)
  return rules.length && rules.letter && rules.digit
}

export function PasswordRules({ value }: PasswordRulesProps) {
  const rules = checkPasswordRules(value)

  return (
    <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
      <p className="text-[11px] font-semibold uppercase tracking-wide text-slate-500">Parol qoidalari</p>
      <ul className="mt-1.5 space-y-1">
        <Rule ok={rules.length} text="10–72 belgi oralig‘ida" />
        <Rule ok={rules.letter} text="Kamida bitta harf" />
        <Rule ok={rules.digit} text="Kamida bitta raqam" />
      </ul>
    </div>
  )
}

function Rule({ ok, text }: { ok: boolean; text: string }) {
  return (
    <li className={`flex items-center gap-2 text-xs ${ok ? 'text-emerald-700' : 'text-slate-500'}`}>
      <span
        className={`flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold ${
          ok ? 'bg-emerald-600 text-white' : 'bg-slate-200 text-slate-500'
        }`}
      >
        {ok ? '✓' : '·'}
      </span>
      {text}
    </li>
  )
}
