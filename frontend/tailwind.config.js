/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#FFFFFF',
          surface: '#F8FAFC',
          card: '#FFFFFF',
          border: '#E2E8F0',
          text: '#0F172A',
          muted: '#64748B'
        },
        accent: {
          blue: '#3B82F6',
          cyan: '#0891B2',
          green: '#10B981',
          amber: '#D97706',
          red: '#EF4444',
          purple: '#8B5CF6',
        }
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', 'monospace'],
      },
      boxShadow: {
        'card': '0 1px 3px rgba(15,23,42,0.06), 0 0 0 1px rgba(15,23,42,0.04)',
        'card-hover': '0 8px 24px rgba(15,23,42,0.08), 0 0 0 1px rgba(59,130,246,0.2)',
        'glow': '0 0 20px rgba(59,130,246,0.15)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    }
  },
  plugins: []
}
