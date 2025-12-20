.PHONY: install dev build lint test preview clean

# Install dependencies using lockfile
install:
	npm ci

# Start dev server (defaults to localhost:5173)
dev:
	npm run dev -- --host

# Type-check and build production assets
build:
	npm run build

# Lint TypeScript/React sources
lint:
	npm run lint

# Run unit tests (Vitest + jsdom)
test:
	npm test

# Preview the production build
preview:
	npm run preview -- --host

# Remove build artifacts
clean:
	rm -rf node_modules dist .vite
