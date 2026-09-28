import { Button, Input, Space } from 'antd'

import { rangeLast7, rangeLastWeek, rangeToday } from '../lib/dateRange'

type DateRangeFilterProps = {
  from: string
  to: string
  onChange: (from: string, to: string) => void
}

export function DateRangeFilter({ from, to, onChange }: DateRangeFilterProps) {
  return (
    <Space wrap size={8}>
      <Input
        type="date"
        value={from}
        aria-label="开始日期"
        style={{ width: 148 }}
        onChange={(e) => onChange(e.target.value, to)}
      />
      <span style={{ color: '#7b8b92' }}>至</span>
      <Input
        type="date"
        value={to}
        aria-label="结束日期"
        style={{ width: 148 }}
        onChange={(e) => onChange(from, e.target.value)}
      />
      <Button.Group>
        <Button
          onClick={() => {
            const [a, b] = rangeToday()
            onChange(a, b)
          }}
        >
          今天
        </Button>
        <Button
          onClick={() => {
            const [a, b] = rangeLast7()
            onChange(a, b)
          }}
        >
          近7天
        </Button>
        <Button
          onClick={() => {
            const [a, b] = rangeLastWeek()
            onChange(a, b)
          }}
        >
          上周
        </Button>
      </Button.Group>
      {from || to ? (
        <Button type="link" onClick={() => onChange('', '')}>
          清除
        </Button>
      ) : null}
    </Space>
  )
}
