/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["**/*.templ"],
	theme: {
		extend: {
			colors: {
				primary: {
					light: "#DF504B",
					DEFAULT: "#AD2831",
					dark: "#800E13",
				},
				secondary: {
					light: "#E9F0FA",
					DEFAULT: "#245ED6",
					dark: "#0D41AA",
				},
			},
		},
	},
	plugins: [],
};
