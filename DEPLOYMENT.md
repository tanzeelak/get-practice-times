# Heroku Deployment Guide

## Prerequisites

1. **Heroku CLI** installed: https://devcenter.heroku.com/articles/heroku-cli
2. **Git** repository initialized
3. **Heroku account** created

## Deployment Steps

### 1. Login to Heroku
```bash
heroku login
```

### 2. Create Heroku App
```bash
heroku create your-app-name
# Or let Heroku generate a name:
heroku create
```

### 3. Set Buildpack (Go)
```bash
heroku buildpacks:set heroku/go
```

### 4. Optional: Add Redis (for caching)
```bash
# Free tier Redis addon
heroku addons:create heroku-redis:mini
```

### 5. Deploy
```bash
git add .
git commit -m "Prepare for Heroku deployment"
git push heroku main
```

### 6. Open App
```bash
heroku open
```

## Environment Variables

The app will automatically detect:
- `PORT` - Set by Heroku
- `REDIS_URL` - Set by Redis addon (if installed)

## Troubleshooting

### Check Logs
```bash
heroku logs --tail
```

### Test Locally
```bash
# Install dependencies
go mod download

# Run locally
go run server.go
```

### Manual Scaling
```bash
# Ensure at least one dyno is running
heroku ps:scale web=1
```

## Production URLs

- **Frontend**: `https://your-app-name.herokuapp.com/`
- **API**: `https://your-app-name.herokuapp.com/api/rehearsals`
- **Health Check**: `https://your-app-name.herokuapp.com/health`

## Notes

- The app works with or without Redis (caching is optional)
- Static files are served from the `/frontend` directory
- API endpoints are prefixed with `/api`
- The app automatically uses Heroku's assigned port