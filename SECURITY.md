# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in CGAP, please report it responsibly to **tushardharaster@gmail.com** instead of using the public issue tracker.

### What to Include
When reporting a security issue, please provide:
- **Description** of the vulnerability
- **Location** in the code (file, line number if possible)
- **Steps to reproduce** the vulnerability
- **Potential impact** (data exposure, privilege escalation, etc.)
- **Suggested fix** (if you have one)

### Response Timeline
- **24 hours**: Initial acknowledgment of your report
- **72 hours**: Assessment and confirmation
- **7-14 days**: Fix development and testing
- **Coordinated disclosure**: We'll work with you on timing

We appreciate responsible disclosure and will acknowledge your contribution in the security fix release notes (with your permission).

---

## Security Best Practices for Development

### Code Security
- ✅ **Never commit secrets**: API keys, tokens, passwords, or credentials
- ✅ **Use environment variables**: Store all sensitive data in `.env` (which is in `.gitignore`)
- ✅ **Review before pushing**: Use `git diff` to verify your changes
- ✅ **Sign commits**: Use GPG keys for signing commits
- ✅ **Keep dependencies updated**: Regularly run `go get -u` and `npm audit`

### Secret Management
- ✅ **Use `.env.example`**: Template for developers (no real credentials)
- ✅ **Rotate credentials**: Regularly rotate API keys, tokens, and passwords
- ✅ **Least privilege**: Give services minimum required permissions
- ✅ **Audit logging**: Enable logs for sensitive operations
- ✅ **Vault for production**: Use managed services (AWS Secrets Manager, HashiCorp Vault, etc.)

### Deployment Security
- ✅ **Environment variables**: Pass secrets via environment, never in config files
- ✅ **HTTPS everywhere**: Use TLS/SSL for all communications
- ✅ **Rate limiting**: Protect endpoints from abuse
- ✅ **Input validation**: Validate all user inputs
- ✅ **SQL injection prevention**: Use parameterized queries (already done with pgx)
- ✅ **CORS policy**: Restrict cross-origin requests appropriately

### Dependency Security
```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies safely
go get -u -t ./...
go mod tidy

# Test after updates
go test ./...
go build ./...
```

### Container Security
- ✅ **Scan images**: Use Trivy or similar for Docker images
- ✅ **Minimal base images**: Use alpine or distroless images
- ✅ **No root**: Run containers as non-root user
- ✅ **Read-only filesystem**: Where possible

---

## Known Security Considerations

### Authentication
- Currently API key based (project_id + implicit auth)
- For production: Implement OAuth2 or JWT tokens
- Consider: Rate limiting per API key

### Data Protection
- ✅ Database passwords in docker-compose are development only
- Production: Use managed database services with encryption at rest
- PII: Implement data classification and retention policies

### API Security
- ✅ Input validation on all endpoints
- Consider: API rate limiting
- Consider: IP whitelisting for internal endpoints
- Consider: CORS policy enforcement

---

## Security Headers & Best Practices

### Recommended Security Headers
```go
// In your API handler
c.Set("X-Content-Type-Options", "nosniff")
c.Set("X-Frame-Options", "DENY")
c.Set("X-XSS-Protection", "1; mode=block")
c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

### OWASP Top 10 Checklist
- [ ] A01:2021 – Broken Access Control
- [ ] A02:2021 – Cryptographic Failures
- [ ] A03:2021 – Injection
- [ ] A04:2021 – Insecure Design
- [ ] A05:2021 – Security Misconfiguration
- [ ] A06:2021 – Vulnerable and Outdated Components
- [ ] A07:2021 – Identification and Authentication Failures
- [ ] A08:2021 – Software and Data Integrity Failures
- [ ] A09:2021 – Logging and Monitoring Failures
- [ ] A10:2021 – Server-Side Request Forgery (SSRF)

---

## Compliance & Standards

- **OWASP**: Following OWASP Top 10 security practices
- **CWE**: Addressing common weakness enumeration issues
- **GDPR**: Preparing for data protection compliance
- **SOC 2**: Implementing security controls

---

## Support

For security-related questions (not vulnerabilities):
- Review this file
- Check [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- Refer to [OWASP Guide](https://owasp.org/)

For urgent security issues:
- Email: **tushardharaster@gmail.com**
- Response time: 24 hours

---

**Last Updated**: December 23, 2025  
**Status**: Active and maintained
