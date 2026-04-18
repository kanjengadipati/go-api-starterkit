# Security Policy

## Supported Scope

This repository is a public backend foundation project. Security reports are especially welcome for:

- authentication and token handling
- authorization and permission checks
- password reset and email verification flows
- secret handling and deployment configuration
- dependency risks that directly affect runtime behavior

## Reporting a Vulnerability

Please do **not** open a public issue for sensitive vulnerabilities.

Instead, report privately by contacting the maintainer first and include:

- a clear description of the issue
- affected component or file
- reproduction steps or proof of concept
- impact assessment if known
- suggested remediation if you have one

If a private reporting channel is later added, this file should be updated to point to it directly.

## Disclosure Expectations

- Give maintainers reasonable time to investigate and patch the issue before public disclosure.
- Avoid publishing working exploit details while a fix is still being prepared.
- If the issue involves exposed credentials, assume they must be rotated immediately.

## Operational Notes

- Never commit real `.env` values
- Use platform-managed secrets for production
- Rotate any credential that has ever been exposed in git history, screenshots, logs, or config output
