# Frontend Implementation Checklist

## Backend Configuration
- [x] Implement CORS Middleware (`internal/api/middleware/cors.go`)
- [x] Configure Static File Server in `main.go` for `/` route
- [x] Connect Routes: `/` -> `frontend/`, `/v1/` -> API

## Frontend Architecture (Vanilla JS + CSS)
- [x] Create shared CSS (`frontend/css/style.css`) - Dark/Hacker Theme
- [x] Create shared JS (`frontend/js/api.js`) - Fetch Logic
- [x] Create `index.html` (The Portal - "Choose your Reality")

## Specific Views (The Magic)
- [x] `twitter.html` (Micro-blog: Text only)
- [x] `tiktok.html` (Vertical Video: Video only, scroll snap)
- [x] `youtube.html` (Video Player: Grid layout)
- [x] `tabnews.html` (Tech News: Dense text list)
- [x] `facebook.html` (Social Mix: No filters)
