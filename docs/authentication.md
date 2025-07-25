# Authentication System Documentation

## Overview

The A5C Hub authentication system provides enterprise-grade authentication with comprehensive security features including:

- **Multi-Factor Authentication (MFA)** - TOTP, SMS, and WebAuthn support
- **Single Sign-On (SSO)** - SAML 2.0, OIDC, and OAuth 2.0
- **Directory Integration** - LDAP and Active Directory support
- **Session Management** - Advanced session security and monitoring
- **Security Features** - Rate limiting, audit logging, and threat detection

## Features

### Core Authentication
- User registration and login
- Email verification
- Password reset functionality
- Secure password policies
- Account lockout protection

### Multi-Factor Authentication
- **TOTP** - Time-based One-Time Passwords (Google Authenticator, Authy)
- **SMS** - SMS-based verification codes
- **WebAuthn** - Hardware security keys and biometric authentication
- **Backup Codes** - Recovery codes for account access

### OAuth Providers
- GitHub
- Google
- Microsoft/Azure AD
- GitLab (including self-hosted)
- Account linking and unlinking

### Enterprise SSO
- **SAML 2.0** - Enterprise SAML authentication with group-based organization assignment
- **OIDC** - OpenID Connect support with automatic organization creation from groups
- **LDAP/AD** - Directory service integration with external team synchronization
- **Just-in-time (JIT) Provisioning** - Automatic user and organization provisioning
- **External Team Sync** - Synchronization with LDAP, Active Directory, Okta, and GitHub teams

### Session Management
- **Secure Session Tokens** - JWT-based authentication with configurable expiration
- **Refresh Token Management** - Secure token rotation and lifecycle management  
- **Token Blacklisting** - Secure token revocation for logout and security incidents
- **OAuth State Validation** - CSRF protection for OAuth flows with secure state storage
- **Device Tracking** - Device tracking and naming with location tracking (optional)
- **Idle Timeout Protection** - Configurable session timeouts
- **Concurrent Session Limits** - Multi-device session management
- **Remember Me Functionality** - Extended session support

### Security Features
- **Rate Limiting** - Protection against brute force attacks
- **Audit Logging** - Comprehensive security event logging
- **Threat Detection** - Suspicious activity monitoring
- **Account Lockout** - Automatic lockout after failed attempts
- **Password Policies** - Configurable password requirements

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRATION_HOUR=24

# SMTP Configuration
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@example.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=noreply@example.com
SMTP_USE_TLS=true

# OAuth Providers
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
MICROSOFT_CLIENT_ID=your-microsoft-client-id
MICROSOFT_CLIENT_SECRET=your-microsoft-client-secret
MICROSOFT_TENANT_ID=your-tenant-id
GITLAB_CLIENT_ID=your-gitlab-client-id
GITLAB_CLIENT_SECRET=your-gitlab-client-secret
GITLAB_BASE_URL=https://gitlab.example.com

# LDAP Configuration
LDAP_HOST=ldap.example.com
LDAP_PORT=389
LDAP_BASE_DN=dc=example,dc=com
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD=your-ldap-password
LDAP_USER_FILTER=(uid=%s)
```

### YAML Configuration

```yaml
# config.yaml
jwt:
  secret: "your-jwt-secret"
  expiration_hour: 24

smtp:
  host: "smtp.example.com"
  port: "587"
  username: "noreply@example.com"
  password: "your-smtp-password"
  from: "noreply@example.com"
  use_tls: true

oauth:
  github:
    client_id: "your-github-client-id"
    client_secret: "your-github-client-secret"
  google:
    client_id: "your-google-client-id"
    client_secret: "your-google-client-secret"
  microsoft:
    client_id: "your-microsoft-client-id"
    client_secret: "your-microsoft-client-secret"
    tenant_id: "your-tenant-id"
  gitlab:
    client_id: "your-gitlab-client-id"
    client_secret: "your-gitlab-client-secret"
    base_url: "https://gitlab.example.com"

saml:
  enabled: true
  entity_id: "https://your-app.com"
  sso_url: "https://idp.example.com/sso"
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
  attribute_map:
    email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
    name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"

ldap:
  enabled: true
  host: "ldap.example.com"
  port: 389
  base_dn: "dc=example,dc=com"
  bind_dn: "cn=admin,dc=example,dc=com"
  bind_password: "your-ldap-password"
  user_filter: "(uid=%s)"
  attribute_map:
    email: "mail"
    first_name: "givenName"
    last_name: "sn"
    display_name: "displayName"
```

## API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Refresh access token
- `GET /auth/me` - Get current user

### Email Verification
- `POST /auth/verify-email` - Verify email with token
- `POST /auth/resend-verification` - Resend verification email

### Password Management
- `POST /auth/forgot-password` - Request password reset
- `POST /auth/reset-password` - Reset password with token
- `POST /auth/change-password` - Change password (authenticated)

### Multi-Factor Authentication
- `POST /auth/mfa/setup` - Setup TOTP MFA
- `POST /auth/mfa/verify` - Verify TOTP setup
- `POST /auth/mfa/disable` - Disable MFA
- `GET /auth/mfa/backup-codes` - Get backup codes
- `POST /auth/mfa/regenerate-backup-codes` - Regenerate backup codes
- `POST /auth/mfa/sms/send` - Send SMS verification code
- `POST /auth/mfa/webauthn/register/begin` - Begin WebAuthn registration
- `POST /auth/mfa/webauthn/register/finish` - Complete WebAuthn registration

### OAuth
- `GET /auth/oauth/{provider}` - Initiate OAuth flow
- `GET /auth/oauth/{provider}/callback` - OAuth callback
- `GET /auth/oauth/accounts` - Get linked accounts
- `POST /auth/oauth/{provider}/unlink` - Unlink OAuth account

### SAML
- `GET /auth/saml/login` - Initiate SAML login
- `POST /auth/saml/acs` - SAML assertion consumer service
- `GET /auth/saml/metadata` - SAML metadata

### LDAP
- `POST /auth/ldap/login` - LDAP authentication
- `POST /auth/ldap/test` - Test LDAP connection

### Session Management
- `GET /auth/sessions` - Get user sessions
- `DELETE /auth/sessions/{id}` - Revoke specific session
- `DELETE /auth/sessions/all` - Revoke all sessions

### Security
- `GET /auth/security/events` - Get security events
- `GET /auth/security/audit` - Get audit logs
- `GET /auth/security/metrics` - Get security metrics

## Database Schema

### Core Tables

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    location VARCHAR(255),
    website VARCHAR(255),
    company VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    phone_number VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address INET,
    user_agent VARCHAR(255),
    device_name VARCHAR(255),
    location_info VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    is_remembered BOOLEAN DEFAULT FALSE,
    security_flags INTEGER DEFAULT 0,
    last_used_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### MFA Tables

```sql
-- Backup codes
CREATE TABLE backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    code VARCHAR(255) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- WebAuthn credentials
CREATE TABLE webauthn_credentials (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    credential_id VARCHAR(255) UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    name VARCHAR(255) NOT NULL,
    sign_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- SMS verification codes
CREATE TABLE sms_verification_codes (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    code VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

## Security Best Practices

### Password Policies
- Minimum 12 characters
- Must contain uppercase, lowercase, digits, and special characters
- Password history prevention
- Regular password rotation recommendations

### Session Security
- Secure HTTP-only cookies
- CSRF protection
- Session timeout configuration
- Concurrent session limits
- Device tracking and alerts

### Rate Limiting
- Login attempts: 5 per 15 minutes
- Registration: 3 per hour
- Password reset: 3 per hour
- MFA attempts: 10 per 5 minutes
- General API: 100 per minute

### Audit Logging
All security-relevant events are logged including:
- Login/logout events
- Password changes
- MFA setup/disable
- Account lockouts
- Suspicious activity
- Session management
- Configuration changes

### Data Protection
- Passwords hashed with bcrypt
- Sensitive data encrypted at rest
- Secure token generation
- Regular security audits
- GDPR compliance support

## Deployment Considerations

### Production Setup
1. Use strong JWT secrets (>32 characters)
2. Configure HTTPS/TLS for all endpoints
3. Set up proper email delivery (SMTP/SES)
4. Configure external OAuth providers
5. Implement proper logging and monitoring
6. Set up backup and recovery procedures

### Monitoring
- Track authentication metrics
- Monitor failed login attempts
- Alert on suspicious activities
- Regular security audits
- Performance monitoring

### Backup
- Regular database backups
- Configuration backup
- Secret management
- Disaster recovery planning

## Troubleshooting

### Common Issues

**Email not sending**
- Check SMTP configuration
- Verify firewall settings
- Test email service connectivity

**OAuth not working**
- Verify OAuth provider configuration
- Check callback URLs
- Validate client secrets

**LDAP connection issues**
- Test LDAP connectivity
- Verify bind credentials
- Check search filters

**Session issues**
- Check JWT secret configuration
- Verify database connections
- Review session cleanup

### Debug Mode
Enable debug logging for troubleshooting:

```yaml
log_level: 5  # Debug level
```

### Health Checks
- `GET /health/auth` - Authentication system health
- `GET /health/db` - Database connectivity
- `GET /health/email` - Email service status
- `GET /health/ldap` - LDAP connectivity (if enabled)

## Migration Guide

### From Basic Auth
1. Run database migrations
2. Update configuration
3. Test authentication flows
4. Enable new features gradually
5. Train users on new features

### Existing Users
- Existing users maintain their accounts
- Password reset required for enhanced security
- MFA setup encouraged but optional initially
- Session migration handled automatically

## Support

For issues and questions:
- Check logs for error details
- Review configuration settings
- Consult troubleshooting guide
- Contact system administrators