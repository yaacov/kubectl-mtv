# Migration Hooks

Migration hooks enable custom automation during VM migrations using container images to execute tasks before or after migration.

## Overview

Hooks allow you to:
- **Pre-migration**: Validate systems, backup data, stop services
- **Post-migration**: Verify results, cleanup, configure applications

Hook containers have access to migration context through mounted files:
- **`plan.yml`**: Migration plan details and provider information
- **`workload.yml`**: VM-specific information and current state

## Basic Syntax

```bash
kubectl-mtv create hook <name> --image <container-image> [options]
```

**Required**: Only `--image` is required.

## Parameters

### --image (required)
Container image to execute the hook. Can be Ansible runners, shell scripts, Python apps, or any executable container.

```bash
kubectl-mtv create hook my-hook --image quay.io/ansible/creator-ee:latest
kubectl-mtv create hook shell-hook --image alpine/bash:latest
```

### --playbook (optional)
Ansible playbook content, automatically base64 encoded. Use `@filename` to read from file.

```bash
# Inline playbook
kubectl-mtv create hook ansible-hook \
  --image quay.io/ansible/creator-ee:latest \
  --playbook "---
- name: Stop services
  hosts: localhost
  tasks:
    - debug: msg='Stopping services before migration'"

# From file
kubectl-mtv create hook file-hook \
  --image quay.io/ansible/creator-ee:latest \
  --playbook @my-playbook.yaml
```

### --service-account (optional)
Kubernetes service account for hook execution permissions.

```bash
kubectl-mtv create hook secure-hook \
  --image my-registry/hook:v1.0 \
  --service-account migration-sa
```

### --deadline (optional)
Timeout in seconds for hook execution.

```bash
kubectl-mtv create hook timed-hook \
  --image my-registry/backup:latest \
  --deadline 600  # 10 minutes
```

## Examples

### Database Backup Hook
```bash
kubectl-mtv create hook database-backup \
  --image quay.io/ansible/creator-ee:latest \
  --service-account migration-sa \
  --deadline 300 \
  --playbook @backup-playbook.yaml
```

### Shell Script Hook
```bash
kubectl-mtv create hook cleanup-tasks \
  --image alpine/bash:latest \
  --deadline 120
```

### Python Application Hook
```bash
kubectl-mtv create hook api-integration \
  --image my-registry/python-automation:v1.0 \
  --service-account api-client
```

## Using Hooks in Migration Plans

### With Create Plan Flags (Recommended)

Add hooks to all VMs in a plan using flags:

```bash
# Pre-hook only
kubectl-mtv create plan my-migration \
  --source vmware-provider \
  --target openshift-target \
  --vms vm1,vm2,vm3 \
  --pre-hook database-backup

# Both pre and post hooks
kubectl-mtv create plan full-migration \
  --source vmware-provider \
  --target openshift-target \
  --vms @vm-list.yaml \
  --pre-hook validation-check \
  --post-hook cleanup-tasks
```

### Manual Plan Configuration

For per-VM customization, edit the plan directly:

```yaml
apiVersion: forklift.konveyor.io/v1beta1
kind: Plan
metadata:
  name: custom-plan
spec:
  vms:
    - id: vm-001
      hooks:
        - hook:
            namespace: default
            name: database-backup
          step: PreHook
        - hook:
            namespace: default
            name: cleanup-tasks
          step: PostHook
```

**Note**: For PreHooks, the VM must be started and SSH accessible.

## Hook Management

```bash
# List hooks
kubectl-mtv get hook

# View hook details
kubectl-mtv describe hook my-hook

# Delete hook
kubectl-mtv delete hook my-hook
```

## Example Ansible Playbook

```yaml
---
- name: Migration Hook
  hosts: localhost
  vars_files:
    - plan.yml      # Migration context
    - workload.yml  # VM information
  tasks:
    - name: Get SSH credentials
      k8s_info:
        api_version: v1
        kind: Secret
        name: privkey
        namespace: openshift-mtv
      register: ssh_key

    - name: Setup SSH
      copy:
        dest: ~/.ssh/id_rsa
        content: "{{ ssh_key.resources[0].data.key | b64decode }}"
        mode: 0600

    - name: Connect to VM and stop services
      shell: |
        ssh root@{{ vm.ipaddress }} "systemctl stop mariadb"
```

## Troubleshooting

### Common Issues

**Hook already exists**:
```bash
kubectl delete hook my-hook  # Delete existing hook
```

**Invalid image**:
Verify image accessibility:
```bash
docker pull my-registry/hook-image:latest
```

**Service account not found**:
```bash
kubectl create serviceaccount my-hook-sa
```

**View hook logs during migration**:
```bash
kubectl logs <hook-pod-name>
```

For more examples and advanced usage, see the full kubectl-mtv documentation. 