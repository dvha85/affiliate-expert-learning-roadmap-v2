# Security policy

This repository is an offline-first learning and readiness project. It is
explicitly `NOT_READY_FOR_PRODUCTION`; fixtures and CI do not grant production
authority.

Do not report credentials, personal data, private evidence, or live provider
responses in a public issue. For a suspected vulnerability, contact the
repository owner privately through the GitHub security-advisory workflow (if
enabled) or an account-controlled private channel, and include reproduction
steps, affected commit, and whether a secret may have been exposed.

The loopback canonical adapter requires `CANONICAL_ADAPTER_TOKEN` for all
stateful HTTP endpoints. Never commit that token or a real provider key.
