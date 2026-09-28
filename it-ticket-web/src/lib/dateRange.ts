export function formatDay(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function rangeToday(): [string, string] {
  const t = formatDay(new Date())
  return [t, t]
}

export function rangeLast7(): [string, string] {
  const end = startOfDay(new Date())
  const start = new Date(end)
  start.setDate(end.getDate() - 6)
  return [formatDay(start), formatDay(end)]
}

export function rangeLastWeek(): [string, string] {
  const now = startOfDay(new Date())
  const weekday = now.getDay()
  const daysFromMonday = weekday === 0 ? 6 : weekday - 1
  const thisMonday = new Date(now)
  thisMonday.setDate(now.getDate() - daysFromMonday)
  const lastMonday = new Date(thisMonday)
  lastMonday.setDate(thisMonday.getDate() - 7)
  const lastSunday = new Date(thisMonday)
  lastSunday.setDate(thisMonday.getDate() - 1)
  return [formatDay(lastMonday), formatDay(lastSunday)]
}
