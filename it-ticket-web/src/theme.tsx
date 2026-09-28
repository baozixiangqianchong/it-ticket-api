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
    colorText: '#12202a',
    colorTextSecondary: '#5b6b73',
    colorBorder: '#d7e2de',
    colorBorderSecondary: '#e4eeea',
    borderRadius: 10,
    fontSize: 14,
    controlHeight: 36,
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
      itemMarginInline: 8,
      itemBorderRadius: 8,
    },
    Card: {
      headerFontSize: 15,
      headerHeight: 52,
    },
    Table: {
      headerBg: '#f4f8f6',
      headerColor: '#4b5c63',
      rowHoverBg: '#f3faf7',
    },
    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
    },
    Segmented: {
      itemSelectedBg: '#fff',
      trackBg: '#e4eeea',
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
