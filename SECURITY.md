# Security Policy

## Supported Versions

| Version | Supported |
|---------|------------|
| 0.1.x   | ✅ |
| < 0.1.0  | ❌ |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly.

### How to Report

1. **Do not** create a public issue or pull request
2. Send an email to: security@hf-local-hub.com
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if known)

### What to Expect

- We will acknowledge receipt within 48 hours
- We will provide a detailed response within 7 days
- We will work with you on a timeline for fixing the issue
- We will credit you in the release notes if desired

### Security Update Process

1. Issue is triaged and confirmed
2. Fix is developed in a private branch
3. Security advisory is drafted
4. Fix is deployed and announced
5. Advisory is published

## Vulnerability Types

We prioritize vulnerabilities based on:

- **Critical**: Remote code execution, privilege escalation
- **High**: Data exposure, authentication bypass
- **Medium**: DoS, sensitive data disclosure
- **Low**: Information disclosure, minor impact

## Security Best Practices

### For Users

1. **Do not expose to internet**: hf-local-hub is designed for local use
2. **Use firewalls**: Restrict access to trusted networks only
3. **Regular updates**: Keep updated with latest releases
4. **Backup data**: Regularly backup your model storage
5. **Monitor logs**: Watch for suspicious activity

### For Developers

1. **Input validation**: All inputs must be validated
2. **Path sanitization**: Prevent directory traversal
3. **Error handling**: Don't leak sensitive information
4. **Dependencies**: Keep dependencies updated
5. **Security reviews**: Code changes should be reviewed

## Known Security Considerations

### Authentication
Current version does not implement authentication. All requests are public.

**Recommendation**: Use firewall rules to restrict access.

### Encryption
Files are stored in plain text. No encryption at rest.

**Recommendation**: Use encrypted filesystem for sensitive models.

### Network Security
No TLS/HTTPS support. All traffic is unencrypted.

**Recommendation**: Use reverse proxy with SSL termination for production.

### File Upload
File size limits are not enforced.

**Recommendation**: Configure web server limits if using reverse proxy.

## Dependencies

We regularly update dependencies to address security issues. See:
- Go: `server/go.mod`
- Python: `python/pyproject.toml`

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Contact

For security-related questions:
- Email: security@hf-local-hub.com
- GitHub Security: [Report a vulnerability](https://github.com/lyani/hf-local-hub/security/advisories/new)
