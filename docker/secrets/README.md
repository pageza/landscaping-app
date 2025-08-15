# Secrets Directory

This directory contains sensitive configuration files for production deployment. **Never commit actual secret files to version control.**

## Production Secrets Required

Create the following files with actual secret values:

- `postgres_password.txt` - PostgreSQL database password
- `redis_password.txt` - Redis authentication password
- `jwt_secret.txt` - JWT token signing secret (minimum 64 characters)
- `encryption_key.txt` - Data encryption key (exactly 32 characters)
- `session_secret.txt` - Web session signing secret

## Generating Secrets

Use these commands to generate secure secrets:

```bash
# PostgreSQL password (32 characters)
openssl rand -base64 32 > postgres_password.txt

# Redis password (32 characters)
openssl rand -base64 32 > redis_password.txt

# JWT secret (64 characters)
openssl rand -base64 64 > jwt_secret.txt

# Encryption key (exactly 32 characters)
openssl rand -base64 32 | cut -c1-32 > encryption_key.txt

# Session secret (64 characters)
openssl rand -base64 64 > session_secret.txt
```

## File Permissions

Set restrictive permissions on secret files:

```bash
chmod 600 *.txt
```

## Backup

Ensure secrets are backed up securely and separately from the application code.

## Environment-Specific Secrets

- Development: Use simple, non-secure values defined in `.env.dev`
- Staging: Use secure values but can be less restrictive
- Production: Use highly secure, randomly generated values