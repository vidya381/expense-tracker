# Deployment Guide

## Environment Variables Setup

### Backend (Render)

Required environment variables for the Go backend:

**Option 1 (Recommended):** Use DATABASE_URL - simpler and standard for Render
```bash
# Database Configuration - Single connection string from Supabase
# IMPORTANT: Add prefer_simple_protocol=true to avoid prepared statement errors with connection pooling
DATABASE_URL=postgresql://postgres.projectref:password@db.example.supabase.co:6543/postgres?sslmode=require&prefer_simple_protocol=true

# JWT Secret (generate a new one for production)
JWT_SECRET=generate_with_openssl_rand_hex_32

# CORS Configuration
CORS_ORIGIN=https://your-app.vercel.app  # Your Vercel frontend URL
```

**Option 2:** Individual variables (for manual configuration)
```bash
DB_HOST=db.example.supabase.co
DB_PORT=6543  # Use 6543 for connection pooling, 5432 for direct connection
DB_USER=postgres.projectref  # IMPORTANT: Include the project reference (postgres.abcdefgh)
DB_PASSWORD=your_supabase_password
DB_NAME=postgres
DB_SSLMODE=require

JWT_SECRET=generate_with_openssl_rand_hex_32
CORS_ORIGIN=https://your-app.vercel.app
```

**Generate JWT Secret:**
```bash
openssl rand -hex 32
```

### Frontend (Vercel)

Required environment variables for the Next.js frontend:

```bash
NEXT_PUBLIC_API_URL=https://your-backend.onrender.com
```

## Deployment Steps

### 1. Database Setup (Supabase)

1. Create a new project on Supabase
2. Get connection details from: Settings → Database → Connection String
3. Note: Use the "Connection Pooling" string for better performance
4. Run migrations:
   ```bash
   # Set environment variables
   export DB_HOST=your-supabase-host
   export DB_PORT=6543
   export DB_USER=postgres
   export DB_PASSWORD=your-password
   export DB_NAME=postgres

   # Run migration tool
   go run migrate.go
   ```

### 2. Backend Deployment (Render)

1. Create a new Web Service on Render
2. Connect your GitHub repository
3. Configure:
   - **Build Command:** `go build -o expense-tracker-backend`
   - **Start Command:** `./expense-tracker-backend`
   - **Environment:** Go
4. Add all environment variables listed above in the Environment section
5. Deploy

### 3. Frontend Deployment (Vercel)

1. Import project from GitHub on Vercel
2. Framework Preset: Next.js
3. Add environment variable:
   - `NEXT_PUBLIC_API_URL` → Your Render backend URL
4. Deploy

### 4. Post-Deployment Verification

- [ ] Test user registration
- [ ] Test user login
- [ ] Test adding transactions
- [ ] Test viewing dashboard
- [ ] Verify HTTPS is working
- [ ] Check CORS headers allow frontend → backend communication

## Security Checklist

- [ ] `.env` files are NOT committed to git
- [ ] Strong JWT secret generated (32+ characters)
- [ ] Database SSL/TLS enabled (`DB_SSLMODE=require`)
- [ ] CORS_ORIGIN set to exact frontend URL (no wildcards)
- [ ] All environment variables use production values
- [ ] Database credentials rotated from any exposed values

## Local Development Setup

1. Copy `.env.example` to `.env` in both frontend and backend directories
2. Update with your local PostgreSQL credentials
3. Set `DB_SSLMODE=disable` for local PostgreSQL
4. Run backend: `go run main.go db.go`
5. Run frontend: `npm run dev`

## Troubleshooting

### Backend can't connect to database
- Verify `DB_SSLMODE=require` for Supabase
- Check connection string format
- Ensure Supabase project is active

### CORS errors on frontend
- Verify `CORS_ORIGIN` matches your Vercel URL exactly
- Check for trailing slashes (should not have one)
- Ensure backend is deployed and accessible

### Frontend can't reach backend
- Verify `NEXT_PUBLIC_API_URL` is set correctly
- Check backend is running (visit backend URL in browser)
- Ensure HTTPS is working on both frontend and backend
