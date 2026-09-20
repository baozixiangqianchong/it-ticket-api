import { App as AntdApp, ConfigProvider, type ThemeConfig } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import type { ReactNode } from 'react'

export const appTheme: ThemeConfig = {
  token: {
    colorPrimary: '#0f766e',
    colorInfo: '#0f766e',
    colorSuccess: '#15803d',
    colorWarning: '#c2410c',
    colorBgLayout: '#eef4f2',
    borderRadius: 10,
    fontFamily:
      "'PingFang SC', 'Hiragino Sans GB', 'Noto Sans SC', 'Microsoft YaHei', sans-serif",
  },
  components: {
    Layout: {
      siderBg: '#102027',
      headerBg: '#ffffff',
      headerHeight: 64,
      headerPadding: '0 24px',
    },
    Menu: {
      darkItemBg: '#102027',
      darkSubMenuItemBg: '#102027',
      darkItemSelectedBg: '#134e4a',
      darkItemHoverBg: '#16333a',
    },
    Card: {
      headerFontSize: 16,
    },
  },
}

export function AppThemeProvider({ children }: { children: ReactNode }) {
  return (
    <ConfigProvider locale={zhCN} theme={appTheme}>
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  )
}
