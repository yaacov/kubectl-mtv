# Ansible Automation Platform (AAP) API Token Setup

This document describes how to retrieve the AAP URL and create an API token for use with kubectl-mtv using `curl`.

## Retrieving AAP URL from Cluster

The AAP URL can be found via the route in the awx namespace:

```bash
kubectl get routes -n awx
```

Example output:
```text
NAME   HOST/PORT                      PATH   SERVICES      PORT   TERMINATION     WILDCARD
awx    awx-awx.apps.example.com              awx-service   http   edge/Redirect   None
```

The full URL is: `https://<HOST_PORT>`

Get it directly and set as environment variable:
```bash
AAP_URL="https://$(kubectl get route awx -n awx -o jsonpath='{.spec.host}')"
echo "AAP URL: ${AAP_URL}"
```

## Creating an API Token with curl

### Step 1: Get Admin Password

Retrieve the admin password from the Kubernetes secret:

```bash
AAP_PASSWORD=$(kubectl get secret awx-admin-password -n awx -o jsonpath='{.data.password}' | base64 -d)
# The password is now stored in AAP_PASSWORD — do not echo it to the terminal.
```

### Step 2: Set the AAP URL

```bash
export AAP_URL="https://$(kubectl get route awx -n awx -o jsonpath='{.spec.host}')"
```

### Step 3: Create API Token

Create a new API token with write scope and capture it into an environment variable:

```bash
AAP_TOKEN=$(curl -sk -X POST "${AAP_URL}/api/v2/tokens/" \
  -u "admin:${AAP_PASSWORD}" \
  -H "Content-Type: application/json" \
  -d '{"scope": "write", "description": "kubectl-mtv token"}' \
  | jq -r '.token')
```

## Verifying the Token

Test that the token works by querying the AAP API:

```bash
curl -sk -H "Authorization: Bearer ${AAP_TOKEN}" \
  "${AAP_URL}/api/v2/me/"
```

## Configuring AAP in MTV

### Step 1: Create AAP Token Secret

Create a Kubernetes secret with the AAP token:

```bash
kubectl create secret generic aap-token \
  --from-literal=token="${AAP_TOKEN}" \
  -n konveyor-forklift
```

### Step 2: Configure MTV Settings

Set the AAP URL and token secret in the ForkliftController:

```bash
# Get the AAP URL (use internal cluster service for in-cluster access)
AAP_URL="http://awx-service.awx.svc.cluster.local"

# Set AAP URL
kubectl mtv settings set --setting aap_url --value "${AAP_URL}" -n konveyor-forklift

# Set AAP token secret name
kubectl mtv settings set --setting aap_token_secret_name --value "aap-token" -n konveyor-forklift
```

**Note:** Use the internal cluster service URL (`http://awx-service.awx.svc.cluster.local`) for the inventory service running inside the cluster, not the external ingress URL.

### Step 3: Verify Configuration

Check that AAP settings are configured:

```bash
kubectl mtv settings --all | grep aap
```

Expected output:
```text
aap           aap_timeout                                    (not set)                                     0                              
aap           aap_token_secret_name                          aap-token                                     -                              
aap           aap_url                                        http://awx-service.awx.svc.cluster.local      -
```

## Using AAP Job Templates

### List Job Templates

Retrieve the list of available job templates from AAP:

```bash
kubectl mtv get inventory job-template -n konveyor-forklift
```

Example output:
```text
ID  NAME               
───────────────────────
7   Demo Job Template
```

### Create a Migration Hook

Create a hook that triggers an AAP job template during VM migration:

```bash
kubectl mtv create hook \
  --name my-aap-hook \
  --aap-job-template-id 7 \
  -n konveyor-forklift
```

Replace `7` with the job template ID from the `get inventory job-template` output.

### Verify Hook Creation

List the created hooks:

```bash
kubectl get hooks -n konveyor-forklift
```

Get details about your hook:

```bash
kubectl get hook my-aap-hook -n konveyor-forklift -o yaml
```

## Testing AAP Integration

### Test with curl

List inventories:

```bash
curl -sk -H "Authorization: Bearer ${AAP_TOKEN}" \
  "${AAP_URL}/api/v2/inventories/"
```

List job templates:

```bash
curl -sk -H "Authorization: Bearer ${AAP_TOKEN}" \
  "${AAP_URL}/api/v2/job_templates/"
```

## Security Notes

- Tokens provide full API access based on their scope
- Use `write` scope for full read/write operations
- Store tokens securely in Kubernetes secrets
- Rotate tokens periodically
- Delete unused tokens
