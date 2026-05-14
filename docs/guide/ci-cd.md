# CI / CD Integration

EnvGuard is designed for CI/CD pipelines. It exits with clear codes and supports machine-readable output formats.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed / no secrets found |
| `1` | Validation failed or secrets detected |
| `2` | I/O or schema parsing error |

## GitHub Actions

Use the official action:

```yaml
- name: Validate Environment Variables
  uses: firasmosbehi/envguard@v2
  with:
    schema: envguard.yaml
    env: .env
    strict: true
    format: github
```

See [GitHub Action](./github-action) for full details.

## GitLab CI

```yaml
validate-env:
  image: ghcr.io/firasmosbehi/envguard:latest
  script:
    - envguard validate --strict --format json
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

## CircleCI

```yaml
version: 2.1
jobs:
  validate-env:
    docker:
      - image: ghcr.io/firasmosbehi/envguard:latest
    steps:
      - checkout
      - run:
          name: Validate .env
          command: envguard validate --strict
```

## Azure Pipelines

```yaml
steps:
  - script: |
      curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-linux-amd64 -o /tmp/envguard
      chmod +x /tmp/envguard
      /tmp/envguard validate --strict
    displayName: 'Validate Environment Variables'
```

## Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Validate Env') {
            steps {
                sh '''
                    curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-linux-amd64 -o envguard
                    chmod +x envguard
                    ./envguard validate --strict
                '''
            }
        }
    }
}
```

## SARIF Output

For GitHub Advanced Security integration:

```bash
envguard validate --format sarif > envguard-results.sarif
```

Upload to GitHub:

```yaml
- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: envguard-results.sarif
```

## Secret Scanning in CI

```yaml
- name: Scan for Secrets
  uses: firasmosbehi/envguard@v2
  with:
    format: sarif
    scan-secrets: true
```
