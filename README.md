# vaultlink

> A CLI tool to sync and audit HashiCorp Vault secrets across multiple environments

---

## Installation

```bash
go install github.com/yourusername/vaultlink@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultlink.git
cd vaultlink
go build -o vaultlink .
```

---

## Usage

Set your Vault address and token, then run:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.yourtoken"

# Sync secrets from staging to production
vaultlink sync --src staging --dst production --path secret/app

# Audit differences between environments
vaultlink audit --envs staging,production --path secret/app

# List all secrets in an environment
vaultlink list --env staging --path secret/
```

### Common Flags

| Flag | Description |
|------|-------------|
| `--src` | Source environment |
| `--dst` | Destination environment |
| `--path` | Vault secret path |
| `--dry-run` | Preview changes without applying |
| `--config` | Path to config file (default: `~/.vaultlink.yaml`) |

---

## Configuration

```yaml
# ~/.vaultlink.yaml
environments:
  staging:
    addr: https://vault-staging.example.com
  production:
    addr: https://vault-prod.example.com
```

---

## License

This project is licensed under the [MIT License](LICENSE).