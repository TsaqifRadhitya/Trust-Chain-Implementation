/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: '#0B0F19',
        surface: '#1A233A',
        surfaceHover: '#2A3655',
        primary: '#0ea5e9', // Cyber blue
        accent: '#10b981',  // Emerald green
        accentHover: '#059669',
        danger: '#ef4444',
        warning: '#f59e0b',
        success: '#10b981',
        text: '#f8fafc',
        textMuted: '#94a3b8',
      },
    },
  },
  plugins: [],
}
