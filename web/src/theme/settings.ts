/** Default theme settings */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: '#0d7fa8',
  themeRadius: 6,
  otherColor: {
    info: '#0e86b8',
    success: '#159a9c',
    warning: '#b67a20',
    error: '#b94a48'
  },
  isInfoFollowPrimary: true,
  layout: {
    mode: 'vertical',
    scrollMode: 'content'
  },
  page: {
    animate: true,
    animateMode: 'fade-slide'
  },
  header: {
    height: 56,
    breadcrumb: {
      visible: true,
      showIcon: true
    },
    multilingual: {
      visible: true
    },
    globalSearch: {
      visible: true
    }
  },
  tab: {
    visible: true,
    cache: true,
    height: 44,
    mode: 'chrome',
    closeTabByMiddleClick: false
  },
  fixedHeaderAndTab: true,
  sider: {
    inverted: false,
    width: 220,
    collapsedWidth: 64,
    mixWidth: 90,
    mixCollapsedWidth: 64,
    mixChildMenuWidth: 200,
    autoSelectFirstMenu: false
  },
  footer: {
    visible: true,
    fixed: false,
    height: 48,
    right: true
  },
  watermark: {
    visible: false,
    text: 'FootStat 足球数据分析',
    enableUserName: false,
    enableTime: false,
    timeFormat: 'YYYY-MM-DD HH:mm'
  },
  tokens: {
    light: {
      colors: {
        container: 'rgb(255, 255, 255)',
        layout: 'rgb(243, 248, 254)',
        inverted: 'rgb(18, 44, 74)',
        'base-text': 'rgb(18, 38, 58)'
      },
      boxShadow: {
        header: '0 1px 2px rgb(22, 62, 105, 0.08)',
        sider: '2px 0 8px 0 rgb(22, 62, 105, 0.06)',
        tab: '0 1px 2px rgb(22, 62, 105, 0.06)'
      }
    },
    dark: {
      colors: {
        container: 'rgb(28, 28, 28)',
        layout: 'rgb(18, 18, 18)',
        'base-text': 'rgb(224, 224, 224)'
      }
    }
  }
};

/**
 * Override theme settings
 *
 * If publish new version, use `overrideThemeSettings` to override certain theme settings
 */
export const overrideThemeSettings: Partial<App.Theme.ThemeSetting> = {
  themeColor: '#0d7fa8',
  otherColor: {
    info: '#0e86b8',
    success: '#159a9c',
    warning: '#b67a20',
    error: '#b94a48'
  },
  tokens: {
    light: {
      colors: {
        container: 'rgb(255, 255, 255)',
        layout: 'rgb(243, 248, 254)',
        inverted: 'rgb(18, 44, 74)',
        'base-text': 'rgb(18, 38, 58)'
      },
      boxShadow: {
        header: '0 1px 2px rgb(22, 62, 105, 0.08)',
        sider: '2px 0 8px 0 rgb(22, 62, 105, 0.06)',
        tab: '0 1px 2px rgb(22, 62, 105, 0.06)'
      }
    }
  }
};
