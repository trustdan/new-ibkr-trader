/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      fontFamily: {
        'nunito': ['Nunito', 'sans-serif'],
      }
    },
  },
  plugins: [
    require('daisyui'),
    require('@tailwindcss/forms')
  ],
  daisyui: {
    themes: [
      {
        'ibkr-dark': {
          'primary': '#3b82f6',
          'secondary': '#6366f1', 
          'accent': '#06b6d4',
          'neutral': '#374151',
          'base-100': '#1f2937',
          'base-200': '#374151',
          'base-300': '#4b5563',
          'info': '#0ea5e9',
          'success': '#22c55e',
          'warning': '#f59e0b',
          'error': '#ef4444',
        }
      },
      'dark'
    ],
    darkTheme: 'ibkr-dark',
    base: true,
    styled: true,
    utils: true,
  },
} 