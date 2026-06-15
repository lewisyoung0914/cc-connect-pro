import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: {
          DEFAULT: '#FFFFFF',
          dark: '#1C1C1E',
        },
        surface: {
          DEFAULT: '#F5F5F7',
          dark: '#2C2C2E',
        },
        accent: {
          DEFAULT: '#5856D6',
          dark: '#7B79E0',
        },
        success: {
          DEFAULT: '#34C759',
          dark: '#30D158',
        },
        warning: {
          DEFAULT: '#FF9500',
          dark: '#FF9F0A',
        },
        primary: {
          DEFAULT: '#1D1D1F',
          dark: '#F5F5F7',
        },
        secondary: {
          DEFAULT: '#86868B',
          dark: '#A1A1A6',
        },
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Roboto',
          '"Helvetica Neue"',
          'Arial',
          'sans-serif',
        ],
        mono: [
          '"SF Mono"',
          'Menlo',
          'Monaco',
          '"Courier New"',
          'monospace',
        ],
      },
      fontSize: {
        'large-title': ['28px', { lineHeight: '34px' }],
        'title': ['20px', { lineHeight: '25px' }],
        'headline': ['17px', { lineHeight: '22px', fontWeight: '600' }],
        'body': ['15px', { lineHeight: '20px' }],
        'caption': ['12px', { lineHeight: '16px' }],
        'mini': ['10px', { lineHeight: '13px' }],
      },
      borderRadius: {
        DEFAULT: '8px',
        sm: '6px',
      },
      spacing: {
        sidebar: '220px',
        content: '32px',
        card: '16px',
      },
    },
  },
  plugins: [],
} satisfies Config
