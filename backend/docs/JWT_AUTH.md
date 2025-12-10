# JWT Authentication Guide

## Overview

The HKERS backend now uses **JWT (JSON Web Tokens)** for API authentication instead of session cookies. This provides a stateless, scalable authentication mechanism that works well with modern frontend applications.

## Architecture

### Authentication Flow

```
1. User visits frontend
2. Frontend redirects to /auth/login
3. Backend redirects to OIDC provider (Auth0, Google, etc.)
4. User logs in with OIDC provider
5. OIDC provider redirects to /auth/callback
6. Backend verifies OIDC token, validates user, generates JWT
7. Backend returns JWT in response body (not cookie!)
8. Frontend stores JWT (localStorage/sessionStorage)
9. Frontend includes JWT in Authorization header for all API requests
```

### Key Changes from Session-Based Auth

| Aspect | Session-Based (Old) | JWT-Based (New) |
|--------|---------------------|-----------------|
| **Storage** | Server (Redis) | Client (localStorage) |
| **State** | Stateful | Stateless |
| **Scalability** | Requires sticky sessions or shared session store | No special requirements |
| **Mobile Support** | Difficult | Native support |
| **CORS** | Complex (`credentials: true`) | Simple (standard headers) |
| **Transport** | Cookies (automatic) | Authorization header (manual) |

## API Endpoints

### POST /auth/login
Initiates the OIDC login flow. Redirects to OIDC provider.

**Request:**
```http
GET /auth/login HTTP/1.1
```

**Response:**
```
302 Redirect to OIDC provider
```

---

### GET /auth/callback
OIDC callback endpoint. Returns JWT token.

**Query Parameters:**
- `code`: Authorization code from OIDC provider
- `state`: State parameter for CSRF protection

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 604800,
    "user": {
      "id": 2,
      "email": "user@example.com",
      "username": "user123",
      "oidc_sub": "google-oauth2|112951580310140109215",
      "is_active": true,
      "trust_points": 0,
      "created_at": "2025-12-10T16:19:13.595Z"
    }
  }
}
```

---

### POST /auth/refresh
Refreshes an existing JWT token.

**Request:**
```http
POST /auth/refresh HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 604800
  }
}
```

---

### POST /auth/logout
Client-side logout. Returns optional OIDC provider logout URL.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully",
    "logout_url": "https://your-oidc-provider.com/v2/logout?..."
  }
}
```

---

### Protected Endpoints
All protected endpoints require a valid JWT token.

**Request:**
```http
GET /api/v1/me HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response (Unauthorized):**
```json
{
  "success": false,
  "error": "Authorization header required"
}
```

## Frontend Integration

### Login Flow

```javascript
// 1. Redirect to login endpoint
window.location.href = 'http://localhost:3000/auth/login';

// 2. After OIDC callback, backend returns to your frontend with token
// You need to capture this in your callback route
// Example: http://your-frontend.com/callback?token=eyJ...

// 3. Store token
localStorage.setItem('access_token', tokenFromCallback);
```

### Making Authenticated Requests

```javascript
// Fetch example
const response = await fetch('http://localhost:3000/api/v1/me', {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
    'Content-Type': 'application/json'
  }
});

// Axios example
axios.defaults.headers.common['Authorization'] = 
  `Bearer ${localStorage.getItem('access_token')}`;

const response = await axios.get('http://localhost:3000/api/v1/me');
```

### Token Refresh

```javascript
async function refreshToken() {
  try {
    const response = await fetch('http://localhost:3000/auth/refresh', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      }
    });
    
    const data = await response.json();
    
    if (data.success) {
      localStorage.setItem('access_token', data.data.access_token);
      return true;
    }
  } catch (error) {
    console.error('Token refresh failed:', error);
    // Redirect to login
    window.location.href = '/login';
  }
  return false;
}
```

### Automatic Token Refresh

```javascript
// Intercept 401 responses and refresh token
axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      const success = await refreshToken();
      if (success) {
        originalRequest.headers['Authorization'] = 
          `Bearer ${localStorage.getItem('access_token')}`;
        return axios(originalRequest);
      }
    }
    
    return Promise.reject(error);
  }
);
```

### Logout

```javascript
async function logout() {
  try {
    const response = await fetch('http://localhost:3000/auth/logout', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      }
    });
    
    const data = await response.json();
    
    // Clear local token
    localStorage.removeItem('access_token');
    
    // Optionally logout from OIDC provider
    if (data.data.logout_url) {
      window.location.href = data.data.logout_url;
    } else {
      window.location.href = '/';
    }
  } catch (error) {
    console.error('Logout failed:', error);
  }
}
```

## Configuration

### Environment Variables

```bash
# JWT secret key (required)
JWT_SECRET=your-secret-key-here

# JWT token duration (optional, default: 168h = 7 days)
JWT_DURATION=168h
```

### Generate JWT Secret

```bash
# Linux/Mac
openssl rand -base64 32

# PowerShell (Windows)
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

## Security Considerations

### Token Storage

**✅ Recommended: localStorage/sessionStorage**
- Simple implementation
- Works across tabs (localStorage only)
- Vulnerable to XSS attacks

**⚠️ Alternative: httpOnly cookies**
- More secure against XSS
- Requires CORS configuration
- More complex implementation

**🚫 Never: Regular cookies accessible via JavaScript**
- Vulnerable to both XSS and CSRF

### Best Practices

1. **Always use HTTPS in production**
   ```bash
   GIN_MODE=release  # Enables secure cookies
   ```

2. **Set appropriate token expiration**
   ```bash
   JWT_DURATION=24h  # Shorter = more secure, more frequent refreshes
   ```

3. **Validate tokens on every request**
   - Middleware automatically validates JWT signature
   - Checks expiration
   - Verifies user is still active

4. **Handle token refresh proactively**
   - Refresh before expiration (e.g., at 80% of lifetime)
   - Don't wait for 401 errors

5. **Sanitize all user inputs**
   - Prevent XSS attacks that could steal tokens

6. **Use Content Security Policy (CSP) headers**
   ```go
   // Add to your router
   router.Use(func(c *gin.Context) {
       c.Header("Content-Security-Policy", "default-src 'self'")
       c.Next()
   })
   ```

## JWT Token Structure

### Claims

```json
{
  "user_id": 2,
  "email": "user@example.com",
  "oidc_sub": "google-oauth2|112951580310140109215",
  "username": "user123",
  "is_active": true,
  "exp": 1765419554,  // Expiration timestamp
  "iat": 1765383554,  // Issued at timestamp
  "nbf": 1765383554   // Not before timestamp
}
```

### Accessing Claims in Handlers

```go
// In your Gin handler
func MyHandler(ctx *gin.Context) {
    // JWT middleware automatically sets these
    userID, _ := middleware.GetUserIDFromContext(ctx)
    email, _ := middleware.GetEmailFromContext(ctx)
    username, _ := middleware.GetUsernameFromContext(ctx)
    
    // Use the user info
    log.Printf("Request from user %d (%s)", userID, email)
}
```

## Troubleshooting

### "Authorization header required"
- Make sure you're including the `Authorization` header
- Format: `Bearer <token>` (note the space after "Bearer")

### "Invalid or expired token"
- Token has expired, use `/auth/refresh` endpoint
- Token signature is invalid (wrong JWT_SECRET)
- Token was tampered with

### "User account is not active"
- User exists in database but `is_active` is false
- Admin needs to activate the account

### CORS Issues
- Make sure your frontend origin is in `CORS_ALLOW_ORIGINS`
- Don't need `credentials: true` for JWT auth (unlike cookies)

## Migration from Session-Based Auth

If you have existing code using session-based auth:

### Before (Session-based)
```javascript
// Cookies were automatic
fetch('http://localhost:3000/api/v1/me', {
  credentials: 'include'  // Include cookies
});
```

### After (JWT-based)
```javascript
// Manual Authorization header
fetch('http://localhost:3000/api/v1/me', {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
  }
});
```

## Additional Resources

- [JWT.io](https://jwt.io/) - JWT decoder and debugger
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [RFC 7519 - JWT Standard](https://tools.ietf.org/html/rfc7519)

