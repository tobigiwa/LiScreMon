build:
	@rm -rf *_templ.go && templ generate

fmt:
	@templ fmt .

live:
	@go generate && templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v

dev: build fmt live

tailwind:
	@tailwindcss -i assets/styles/input.css  -o assets/styles/output.css --watch
