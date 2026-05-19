/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        heading: ['Instrument Serif', 'serif'],
        body: ['Barlow', 'sans-serif'],
      },
      borderRadius: {
        DEFAULT: '9999px',
      },
      colors: {
        'space': {
          950: '#020105',
          900: '#0a0a10',
          800: '#12121a',
          700: '#1a1a24',
        },
        'glass': {
          light: 'rgba(255, 255, 255, 0.04)',
          medium: 'rgba(255, 255, 255, 0.06)',
          strong: 'rgba(255, 255, 255, 0.10)',
          border: 'rgba(255, 255, 255, 0.08)',
        },
      },
      animation: {
        'fade-in': 'fadeIn 0.6s ease-out forwards',
        'fade-in-up': 'fadeInUp 0.8s ease-out forwards',
        'blur-in': 'blurIn 1s ease-out forwards',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeInUp: {
          '0%': { opacity: '0', transform: 'translateY(24px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        blurIn: {
          '0%': { opacity: '0', filter: 'blur(12px)' },
          '100%': { opacity: '1', filter: 'blur(0px)' },
        },
      },
    },
  },
  corePlugins: {
    preflight: false, // Disable Tailwind reset to avoid conflicts with Ant Design
  },
  plugins: [],
}
