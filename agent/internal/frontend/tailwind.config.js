/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/*.templ"],
  theme: {
    screens: {
      max650: {'max':'650px'},
      max550: {'max':'550px'},
    },
    extend: {
      spacing: {
        '99%': '99%',
        '30vh': '30vh',
        'chart-highlight': 'calc(100vh - 37rem)',
        'task-page': 'calc(100vh - 7rem)',
        'white-line': 'calc(100% - 30px)',
      }
    },
  },
  plugins: [],
}

