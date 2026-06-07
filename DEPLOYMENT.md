# Deployment Cheatsheet

Tasks to publish the monorepo to GitHub and deploy both backend (Go) and frontend (Vite/React). The backend defaults to SQLite for local/demo use. Cloud Run can still run it, but Cloud Run's filesystem is ephemeral, so use Postgres mode or durable storage for data you care about.

## 1) Prepare the repo
- Ensure `backend/.env.example` exists (no secrets); keep `backend/.env` out of git via `.gitignore`.
- `git init`, add all sources, commit; create GitHub repo; `git remote add origin git@github.com:<you>/podcast-summarizer.git`; `git push -u origin main`.

## 2) Backend image (Go)
- Add `backend/Dockerfile` (if missing):
  ```Dockerfile
  FROM golang:1.25 AS build
  WORKDIR /app
  COPY . .
  go build -o server ./cmd/server  # adjust path to your main

  FROM gcr.io/distroless/base-debian12
  WORKDIR /app
  COPY --from=build /app/server .
  EXPOSE 8080
  CMD ["/app/server"]
  ```
- Build and push: `gcloud builds submit --tag gcr.io/$PROJECT_ID/podcast-api ./backend`.

## 3) Frontend image (Vite)
- Add `frontend/Dockerfile` (if missing):
  ```Dockerfile
  FROM node:20 AS build
  WORKDIR /app
  COPY package*.json ./
  npm ci
  COPY . .
  npm run build

  FROM nginx:1.27-alpine
  COPY --from=build /app/dist /usr/share/nginx/html
  EXPOSE 8080
  CMD ["nginx", "-g", "daemon off;"]
  ```
- Build and push: `gcloud builds submit --tag gcr.io/$PROJECT_ID/podcast-web ./frontend`.
*Alternative:* skip Cloud Run for the frontend and upload `frontend/dist` to Cloud Storage + Cloud CDN.

## 4) Deploy to Cloud Run
- Backend:
  ```bash
  gcloud run deploy podcast-api \
    --image gcr.io/$PROJECT_ID/podcast-api \
    --region us-central1 \
    --allow-unauthenticated \
    --set-secrets OPENAI_API_KEY=projects/$PROJECT_ID/secrets/OPENAI_API_KEY:latest \
    --set-env-vars "ENV=prod,STORAGE_DRIVER=postgres" \
    --service-account <sa>@$PROJECT_ID.iam.gserviceaccount.com
  ```
- Frontend:
  ```bash
  gcloud run deploy podcast-web \
    --image gcr.io/$PROJECT_ID/podcast-web \
    --region us-central1 \
    --allow-unauthenticated
  ```

## 5) Secrets, env, networking
- Use Secret Manager + `--set-secrets` for keys; use `--set-env-vars` for non-sensitive config (API base URL, etc.).
- Local/demo mode uses `STORAGE_DRIVER=sqlite`, `SQLITE_PATH=data/podcast.db`, `OBJECT_STORAGE_DRIVER=local`, and `LOCAL_STORAGE_PATH=data/storage`.
- For Cloud Run with durable data, set `STORAGE_DRIVER=postgres` and `POSTGRES_URL`.
- `VALKEY_URL` is required when `STORAGE_DRIVER=postgres`; SQLite mode uses SQLite-backed job locking.
- Use `OBJECT_STORAGE_DRIVER=r2` for Cloud Run if generated audio/chunk artifacts must survive instance restarts. Local storage is best for local/self-hosted demos.
- If using Cloud SQL or Memorystore, add a Serverless VPC connector and `--add-cloudsql-instances` or private IP; set `POSTGRES_URL`/`VALKEY_URL`.
- Enable CORS on the backend if the frontend is on a different origin.

## 6) Optional CI/CD (Cloud Build)
- Add `cloudbuild.yaml` to build and deploy both images on push:
  ```yaml
  steps:
    - name: gcr.io/cloud-builders/docker
      args: ["build","-t","gcr.io/$PROJECT_ID/podcast-api","backend"]
    - name: gcr.io/cloud-builders/docker
      args: ["build","-t","gcr.io/$PROJECT_ID/podcast-web","frontend"]
    - name: gcr.io/cloud-builders/gcloud
      args: ["run","deploy","podcast-api","--image","gcr.io/$PROJECT_ID/podcast-api","--region","us-central1","--allow-unauthenticated"]
    - name: gcr.io/cloud-builders/gcloud
      args: ["run","deploy","podcast-web","--image","gcr.io/$PROJECT_ID/podcast-web","--region","us-central1","--allow-unauthenticated"]
  images:
    - gcr.io/$PROJECT_ID/podcast-api
    - gcr.io/$PROJECT_ID/podcast-web
  ```
- Create a Cloud Build trigger from GitHub on `main` or tags; inject secrets via Secret Manager substitutions.

## 7) Post-deploy checks
- Hit backend health or `/` endpoint; load frontend URL; verify CORS and API base URL.
- Map custom domains to both services; force HTTPS.
- Add uptime checks/alerts; watch Cloud Run logs for errors.
