/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["**/*.templ"],
	darkMode: "selector",
	theme: {
		extend: {
			colors: {
				primary: {
					light: "var(--primary-light)",
					DEFAULT: "var(--primary)",
					dark: "var(--primary-dark)",
				},
				accent: {
					light: "var(--accent-light)",
					DEFAULT: "var(--accent)",
					dark: "var(--accent-dark)",
				},

				body: "var(--body)",
				background: "var(--background)",
				section: "var(--section)",
				border: "var(--border)",

				danger: "var(--danger)",
				success: "var(--success)",
				warning: "var(--warning)",
			},
			accentColor: ({ theme }) => theme("colors.accent"),
			borderColor: ({ theme }) => theme("colors.border"),
			borderRadius: {
				lg: "var(--radius)",
				md: "calc(var(--radius) - 2px)",
				sm: "calc(var(--radius) - 4px)",
			},
			transitionProperty: {
				button: "transform, color, background-color, border-color",
			},
		},
	},
	plugins: [],
};
