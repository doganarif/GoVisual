/** @type {import('tailwindcss').Config} */
export default {
    darkMode: ["class"],
    content: ["./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
  	extend: {
  		colors: {
  			border: 'hsl(214 32% 91%)',
  			input: 'hsl(214 32% 91%)',
  			ring: 'hsl(221 83% 53%)',
  			background: 'hsl(0 0% 100%)',
  			foreground: 'hsl(222 47% 11%)',
  			primary: {
  				DEFAULT: 'hsl(221 83% 53%)',
  				foreground: 'hsl(0 0% 100%)'
  			},
  			secondary: {
  				DEFAULT: 'hsl(214 32% 96%)',
  				foreground: 'hsl(222 47% 11%)'
  			},
  			destructive: {
  				DEFAULT: 'hsl(0 84% 60%)',
  				foreground: 'hsl(0 0% 100%)'
  			},
  			muted: {
  				DEFAULT: 'hsl(214 32% 96%)',
  				foreground: 'hsl(215 14% 45%)'
  			},
  			accent: {
  				DEFAULT: 'hsl(214 32% 96%)',
  				foreground: 'hsl(222 47% 11%)'
  			},
  			popover: {
  				DEFAULT: 'hsl(0 0% 100%)',
  				foreground: 'hsl(222 47% 11%)'
  			},
  			card: {
  				DEFAULT: 'hsl(0 0% 100%)',
  				foreground: 'hsl(222 47% 11%)'
  			},
  			success: {
  				DEFAULT: 'hsl(142 76% 36%)',
  				foreground: 'hsl(0 0% 100%)'
  			},
  			warning: {
  				DEFAULT: 'hsl(45 93% 47%)',
  				foreground: 'hsl(222 47% 11%)'
  			},
  			sidebar: {
  				DEFAULT: 'hsl(var(--sidebar-background))',
  				foreground: 'hsl(var(--sidebar-foreground))',
  				primary: 'hsl(var(--sidebar-primary))',
  				'primary-foreground': 'hsl(var(--sidebar-primary-foreground))',
  				accent: 'hsl(var(--sidebar-accent))',
  				'accent-foreground': 'hsl(var(--sidebar-accent-foreground))',
  				border: 'hsl(var(--sidebar-border))',
  				ring: 'hsl(var(--sidebar-ring))'
  			}
  		},
  		borderRadius: {
  			lg: '0.5rem',
  			md: '0.375rem',
  			sm: '0.25rem'
  		},
  		animation: {
  			'fade-in': 'fadeIn 0.2s ease-in-out',
  			'slide-in': 'slideIn 0.2s ease-out'
  		},
  		keyframes: {
  			fadeIn: {
  				'0%': {
  					opacity: 0
  				},
  				'100%': {
  					opacity: 1
  				}
  			},
  			slideIn: {
  				'0%': {
  					transform: 'translateY(-10px)',
  					opacity: 0
  				},
  				'100%': {
  					transform: 'translateY(0)',
  					opacity: 1
  				}
  			}
  		}
  	}
  },
  plugins: [],
};
