# Deployment Guide

## Security Configuration

### Environment Variables

The backend supports several environment variables for security configuration:

#### HTTPS Enforcement

```bash
# Enable HTTPS enforcement in production
ENFORCE_HTTPS=true
```

When `ENFORCE_HTTPS=true`:
- All HTTP requests are redirected to HTTPS (301 Moved Permanently)
- HSTS (HTTP Strict Transport Security) header is added
- Works with proxies/load balancers (checks X-Forwarded-Proto, X-Forwarded-Ssl headers)

**Note:** Only enable this when your application is served over HTTPS.

#### CORS Configuration

```bash
# Set allowed CORS origin (default: http://localhost:3000)
CORS_ORIGIN=https://your-frontend-domain.com
```

#### Content Security Policy

```bash
# Override default CSP (optional)
CSP_HEADER="default-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
```

Default CSP is strict for JSON API responses. Customize if needed.

#### JWT Secret

```bash
# Required: Set a strong JWT secret
JWT_SECRET=your-strong-secret-here
```

**Important:** Use a cryptographically strong secret in production (minimum 32 characters).

### Security Headers

The following security headers are automatically added to all responses:

- **X-Frame-Options**: DENY (prevents clickjacking)
- **X-Content-Type-Options**: nosniff (prevents MIME sniffing)
- **X-XSS-Protection**: 1; mode=block (enables XSS filter)
- **Referrer-Policy**: strict-origin-when-cross-origin
- **Permissions-Policy**: restricts geolocation, microphone, camera
- **Content-Security-Policy**: strict policy for API responses
- **Strict-Transport-Security**: added when HTTPS is enforced

### Production Deployment Example

```bash
# Set environment variables
export DB_HOST=your-db-host
export DB_PORT=5432
export DB_USER=your-db-user
export DB_PASSWORD=your-db-password
export DB_NAME=myspendo
export JWT_SECRET=your-strong-secret-minimum-32-characters
export CORS_ORIGIN=https://your-frontend.com
export ENFORCE_HTTPS=true

# Run the application
./myspendo-backend
```

### HTTPS Setup with Reverse Proxy

If using a reverse proxy (nginx, Apache, etc.), ensure:

1. The proxy terminates SSL/TLS
2. The proxy sets `X-Forwarded-Proto: https` header
3. Set `ENFORCE_HTTPS=true` to enable HSTS headers

Example nginx configuration:

```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Database Connection Pool

The application is configured with:
- Max open connections: 25
- Max idle connections: 10
- Connection max lifetime: 5 minutes
- Connection max idle time: 2 minutes

These can be adjusted in `main.go` if needed for your production load.

### Rate Limiting

Authentication endpoints (/register, /login) are rate-limited to:
- 5 requests per 60 seconds per IP address

This helps prevent brute force attacks.

### Security Best Practices

1. **Always use HTTPS in production** - Set `ENFORCE_HTTPS=true`
2. **Use strong JWT secrets** - Minimum 32 random characters
3. **Keep dependencies updated** - Run `go get -u` regularly
4. **Enable firewall rules** - Only expose necessary ports
5. **Use environment variables** - Never commit secrets to git
6. **Monitor logs** - Watch for suspicious activity
7. **Regular backups** - Backup your PostgreSQL database
8. **Update PostgreSQL** - Keep database server patched
