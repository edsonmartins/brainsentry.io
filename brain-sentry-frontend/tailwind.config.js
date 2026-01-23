/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // Brain Sentry Brand Colors (Laranja/Cobre)
        "brain-primary": "#E67E50",
        "brain-primary-dark": "#D97642",
        "brain-primary-light": "#F29A6E",
        "brain-accent": "#F59E0B",
        "brain-accent-dark": "#DC7609",
        "brain-gold": "#FBBF24",
        // Semantic colors
        "brain-success": "#10B981",
        "brain-error": "#EF4444",
        // Backgrounds
        "brain-bg-dark": "#1A1D29",
        "brain-bg-darker": "#0F1117",
        "brain-bg-lighter": "#252936",
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      animation: {
        "pulse-slow": "pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "fade-in": "fadeIn 0.5s ease-out",
        "slide-up": "slideUp 0.5s ease-out",
        "gradient-x": "gradientX 3s ease infinite",
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        slideUp: {
          "0%": { transform: "translateY(20px)", opacity: "0" },
          "100%": { transform: "translateY(0)", opacity: "1" },
        },
        gradientX: {
          "0%, 100%": { backgroundPosition: "0% 50%" },
          "50%": { backgroundPosition: "100% 50%" },
        },
      },
      backgroundImage: {
        "gradient-hero": "linear-gradient(135deg, #E67E50 0%, #F59E0B 100%)",
        "gradient-cta": "linear-gradient(135deg, #F59E0B 0%, #EF4444 100%)",
        "gradient-card": "linear-gradient(180deg, #252936 0%, #1A1D29 100%)",
        "gradient-orange": "linear-gradient(135deg, #E67E50 0%, #DC7609 100%)",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}
