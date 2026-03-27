# Error Reference

All error responses from the Agent Vault proxy are JSON with this shape:

```json
{
  "error": "<code>",
  "message": "<details>"
}
```

## Error Codes

| Status | Code | Meaning | Agent Action |
|--------|------|---------|--------------|
| 400 | `bad_request` | Missing target host in URL | Fix the URL -- ensure it follows `{AGENT_VAULT_ADDR}/proxy/{host}/{path}` |
| 401 | `unauthorized` | Missing, invalid, or expired session token | Verify `AGENT_VAULT_SESSION_TOKEN` is set and not expired (tokens last 24 hours) |
| 403 | `forbidden` | No broker rule matches the target host, or session is not vault-scoped | Call `GET /discover` to see allowed hosts |
| 502 | `credential_not_found` | A credential template references a credential that doesn't exist | Tell the user to add the missing credential via `agent-vault credentials set` |
| 502 | `upstream_error` | The target service is unreachable | Retry after a brief delay, or inform the user the service may be down |
