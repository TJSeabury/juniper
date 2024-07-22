generate:
	@echo "Generating..."
	templ generate
	pnpm dlx tailwindcss -i ./public/styles/input.css -o ./public/styles/app.css

watch:
	@echo "Watching..."
	templ generate --watch &
	pnpm dlx tailwindcss -i ./public/styles/input.css -o ./public/styles/app.css --watch

build:
	@echo "Building..."
	templ build
	pnpm dlx tailwindcss -i ./public/styles/input.css -o ./public/styles/app.css --minify
	go build
