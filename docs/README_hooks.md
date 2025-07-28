# Creating Migration Hooks

This guide explains how to create migration hooks using the `kubectl-mtv create hook` command. Migration hooks are Kubernetes resources that enable custom automation and logic execution during migration processes.

## Overview

Migration hooks are powerful tools that allow you to run custom automation at various points during the migration lifecycle. They enable:
- **Pre-migration Validation**: Verify source system readiness before migration starts
- **Custom Automation**: Execute specific business logic during migration phases
- **Post-migration Tasks**: Perform cleanup or configuration after migration completion
- **Integration Points**: Connect with external systems and tools
- **Ansible Playbook Execution**: Run Ansible automation for complex workflows

Hooks use container images to execute custom logic, with optional Ansible playbook support for automation tasks.

## Basic Syntax

```bash
kubectl-mtv create hook <hook-name> \
  --image <container-image-url> \
  [--service-account <service-account>] \
  [--playbook <playbook-content-or-@file>] \
  [--deadline <seconds>]
```

**Required**: Only the `--image` parameter is required. All other parameters are optional.

## Parameters

### Required Parameters

#### --image (required)
Specifies the container image that will execute the hook logic.

```bash
kubectl-mtv create hook my-hook --image nginx:latest
```

### Optional Parameters

#### --service-account
Specifies the Kubernetes service account to use for hook execution. This controls the permissions available to the hook container.

```bash
kubectl-mtv create hook my-hook \
  --image my-registry/hook-image:latest \
  --service-account migration-service-account
```

#### --playbook
Provides Ansible playbook content to be executed by the hook. The playbook content is automatically base64 encoded for storage. Supports two formats:

1. **Inline content**: Provide playbook YAML directly
2. **File reference**: Use `@filename` to read from a file

```bash
# Inline playbook content
kubectl-mtv create hook my-hook \
  --image my-registry/ansible-hook:latest \
  --playbook "---
- name: Migration Validation
  hosts: localhost
  tasks:
    - debug:
        msg: 'Running migration validation'"

# Playbook from file
kubectl-mtv create hook my-hook \
  --image my-registry/ansible-hook:latest \
  --playbook @validation-playbook.yaml
```

#### --deadline
Sets a timeout (in seconds) for hook execution. If the hook doesn't complete within this time, it will be terminated.

```bash
kubectl-mtv create hook my-hook \
  --image my-registry/hook-image:latest \
  --deadline 300  # 5 minutes
```

## Complete Examples

### Minimal Hook

Create a basic hook with only the required image parameter:

```bash
kubectl-mtv create hook basic-validation \
  --image nginx:latest
```

### Hook with Service Account

Create a hook that uses a specific service account for permissions:

```bash
kubectl-mtv create hook privileged-validation \
  --image my-registry/validation-hook:v1.0 \
  --service-account migration-admin
```

### Hook with Timeout

Create a hook with a 10-minute execution deadline:

```bash
kubectl-mtv create hook long-running-task \
  --image my-registry/backup-hook:latest \
  --deadline 600
```

### Hook with Inline Playbook

Create a hook with Ansible playbook content provided directly:

```bash
kubectl-mtv create hook inline-automation \
  --image quay.io/ansible/creator-ee:latest \
  --service-account ansible-runner \
  --playbook "---
- name: Pre-migration Validation
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Check source system status
      uri:
        url: https://source-system.example.com/health
        method: GET
      register: health_check
    
    - name: Validate system is ready
      assert:
        that:
          - health_check.status == 200
        fail_msg: 'Source system is not ready for migration'
    
    - name: Log validation success
      debug:
        msg: 'Pre-migration validation completed successfully'"
```

### Hook with Playbook from File

Create a hook using a playbook stored in a file:

```bash
# First, create your playbook file
cat > pre-migration-checks.yaml << 'EOF'
---
- name: Pre-migration System Checks
  hosts: localhost
  gather_facts: true
  vars:
    required_disk_space: 100  # GB
    
  tasks:
    - name: Check available disk space
      set_fact:
        available_space: "{{ (ansible_mounts | selectattr('mount', 'equalto', '/') | first).size_available / 1024 / 1024 / 1024 }}"
    
    - name: Ensure sufficient disk space
      assert:
        that:
          - available_space | float > required_disk_space
        fail_msg: "Insufficient disk space. Required: {{ required_disk_space }}GB, Available: {{ available_space }}GB"
    
    - name: Verify network connectivity
      wait_for:
        host: target-system.example.com
        port: 443
        timeout: 30
      
    - name: Create migration directory
      file:
        path: /tmp/migration-workspace
        state: directory
        mode: '0755'
    
    - name: Log readiness status
      debug:
        msg: "System is ready for migration. Available space: {{ available_space }}GB"
EOF

# Create the hook using the playbook file
kubectl-mtv create hook pre-migration-validation \
  --image quay.io/ansible/creator-ee:latest \
  --service-account migration-service-account \
  --playbook @pre-migration-checks.yaml \
  --deadline 300
```

### Full Configuration Hook

Create a hook with all parameters specified:

```bash
kubectl-mtv create hook comprehensive-automation \
  --image registry.redhat.io/ubi8/ubi:latest \
  --service-account migration-automation \
  --deadline 900 \
  --playbook @comprehensive-migration-automation.yaml
```

## Advanced Playbook Examples

### Database Backup Hook

```yaml
---
- name: Database Backup Hook
  hosts: localhost
  gather_facts: false
  vars:
    backup_dir: "/tmp/migration-backups"
    timestamp: "{{ ansible_date_time.epoch }}"
    
  tasks:
    - name: Create backup directory
      file:
        path: "{{ backup_dir }}"
        state: directory
        mode: '0755'
    
    - name: Backup source database
      shell: |
        mysqldump -h {{ source_db_host }} -u {{ source_db_user }} -p{{ source_db_password }} \
          {{ source_db_name }} > {{ backup_dir }}/backup-{{ timestamp }}.sql
      environment:
        source_db_host: "{{ lookup('env', 'SOURCE_DB_HOST') }}"
        source_db_user: "{{ lookup('env', 'SOURCE_DB_USER') }}"
        source_db_password: "{{ lookup('env', 'SOURCE_DB_PASSWORD') }}"
        source_db_name: "{{ lookup('env', 'SOURCE_DB_NAME') }}"
    
    - name: Verify backup file
      stat:
        path: "{{ backup_dir }}/backup-{{ timestamp }}.sql"
      register: backup_file
    
    - name: Confirm backup success
      assert:
        that:
          - backup_file.stat.exists
          - backup_file.stat.size > 0
        fail_msg: "Database backup failed or is empty"
    
    - name: Upload backup to storage
      aws_s3:
        bucket: migration-backups
        object: "{{ source_db_name }}/backup-{{ timestamp }}.sql"
        src: "{{ backup_dir }}/backup-{{ timestamp }}.sql"
        mode: put
```

### Network Configuration Hook

```yaml
---
- name: Network Configuration Hook
  hosts: localhost
  gather_facts: false
  
  tasks:
    - name: Validate target network connectivity
      uri:
        url: "https://{{ target_system }}/api/health"
        method: GET
        timeout: 30
      vars:
        target_system: "{{ lookup('env', 'TARGET_SYSTEM_URL') }}"
      
    - name: Configure firewall rules
      shell: |
        iptables -A INPUT -p tcp --dport 443 -j ACCEPT
        iptables -A INPUT -p tcp --dport 80 -j ACCEPT
      become: true
      
    - name: Test application connectivity
      wait_for:
        host: "{{ item.host }}"
        port: "{{ item.port }}"
        timeout: 60
      loop:
        - { host: "database.example.com", port: 5432 }
        - { host: "cache.example.com", port: 6379 }
        - { host: "messaging.example.com", port: 5672 }
        
    - name: Update DNS configuration
      lineinfile:
        path: /etc/hosts
        line: "{{ target_ip }} {{ target_hostname }}"
        create: yes
      vars:
        target_ip: "{{ lookup('env', 'TARGET_IP') }}"
        target_hostname: "{{ lookup('env', 'TARGET_HOSTNAME') }}"
```

## Hook Resource Management

### List Hooks

```bash
# List all hooks
kubectl get hooks

# List hooks with more details
kubectl-mtv get hook

# List hooks in specific namespace
kubectl-mtv get hook -n migration-project
```

### View Hook Details

```bash
# Get hook details including spec and status
kubectl get hook my-hook -o yaml

# Describe hook for status and events
kubectl describe hook my-hook

# View hook spec in JSON
kubectl get hook my-hook -o jsonpath='{.spec}' | jq
```

### Verify Hook Playbook Content

```bash
# Extract and decode the playbook content
kubectl get hook my-hook -o jsonpath='{.spec.playbook}' | base64 -d

# View complete hook specification
kubectl get hook my-hook -o yaml
```

### Delete Hooks

```bash
# Delete hook using kubectl
kubectl delete hook my-hook

# Delete multiple hooks
kubectl delete hook hook1 hook2 hook3

# Delete hook using MTV command
kubectl-mtv delete hook my-hook
```

## Best Practices

### 1. Use Descriptive Hook Names

Choose names that clearly indicate the hook's purpose:

```bash
# Good: descriptive names
kubectl-mtv create hook pre-migration-database-backup --image backup-tool:latest
kubectl-mtv create hook post-migration-cleanup --image cleanup-tool:latest
kubectl-mtv create hook network-validation --image network-checker:latest

# Avoid: generic names
kubectl-mtv create hook hook1 --image some-image:latest
```

### 2. Set Appropriate Deadlines

Configure realistic timeouts based on the expected execution time:

```bash
# Quick validation (5 minutes)
kubectl-mtv create hook quick-validation \
  --image validation-tool:latest \
  --deadline 300

# Long-running backup (30 minutes)
kubectl-mtv create hook database-backup \
  --image backup-tool:latest \
  --deadline 1800

# Complex automation (1 hour)
kubectl-mtv create hook comprehensive-setup \
  --image automation-tool:latest \
  --deadline 3600
```

### 3. Use Service Accounts for Security

Create dedicated service accounts with minimal required permissions:

```bash
# Create service account with specific permissions
kubectl create serviceaccount migration-hook-sa

# Create role with minimal permissions
kubectl create role migration-hook-role \
  --verb=get,list,create,update \
  --resource=configmaps,secrets

# Bind role to service account
kubectl create rolebinding migration-hook-binding \
  --role=migration-hook-role \
  --serviceaccount=default:migration-hook-sa

# Use the service account in hooks
kubectl-mtv create hook secure-hook \
  --image hook-image:latest \
  --service-account migration-hook-sa
```

### 4. Organize Playbooks in Files

For complex automation, use external files:

```bash
# Organize playbooks by function
mkdir -p migration-playbooks/
echo "# Pre-migration tasks" > migration-playbooks/pre-migration.yaml
echo "# Post-migration tasks" > migration-playbooks/post-migration.yaml
echo "# Validation tasks" > migration-playbooks/validation.yaml

# Create hooks using organized playbooks
kubectl-mtv create hook pre-migration-tasks \
  --image ansible-runner:latest \
  --playbook @migration-playbooks/pre-migration.yaml
```

### 5. Version Your Hook Images

Use specific version tags instead of `latest`:

```bash
# Good: specific versions
kubectl-mtv create hook validation-hook \
  --image my-registry/validation-tool:v1.2.3

# Better: with digest for immutability
kubectl-mtv create hook validation-hook \
  --image my-registry/validation-tool@sha256:abc123...

# Avoid: latest tag
kubectl-mtv create hook validation-hook \
  --image my-registry/validation-tool:latest
```

### 6. Test Playbooks Independently

Test your Ansible playbooks before using them in hooks:

```bash
# Test playbook locally first
ansible-playbook -i localhost, validation-playbook.yaml

# Then create the hook
kubectl-mtv create hook tested-validation \
  --image ansible-runner:latest \
  --playbook @validation-playbook.yaml
```

## Namespace-Specific Hook Management

### Create Hooks in Specific Namespaces

```bash
# Create hook in specific namespace
kubectl-mtv create hook namespace-specific-hook \
  --namespace migration-project \
  --image hook-image:latest

# List hooks in specific namespace
kubectl-mtv get hook -n migration-project
```

### Cross-Namespace Service Accounts

```bash
# Use service account from different namespace
kubectl-mtv create hook cross-ns-hook \
  --image hook-image:latest \
  --service-account system:serviceaccount:kube-system:migration-sa
```

## Integration with Migration Plans

Hooks are typically referenced in migration plan configurations. While hook creation is independent, they become active when referenced by plans:

```yaml
# Example migration plan referencing hooks
apiVersion: forklift.konveyor.io/v1beta1
kind: Plan
metadata:
  name: my-migration-plan
spec:
  # ... other plan configuration
  hooks:
    - hook:
        name: pre-migration-validation
        namespace: default
      step: PreHook
    - hook:
        name: post-migration-cleanup  
        namespace: default
      step: PostHook
```

## Troubleshooting

### Common Issues

#### 1. Invalid Image Reference
```bash
Error: failed to create hook my-hook: invalid image reference
```
**Solution**: Ensure the image URL is valid and accessible:
```bash
# Test image accessibility
docker pull my-registry/hook-image:latest

# Use correct image format
kubectl-mtv create hook my-hook --image my-registry/hook-image:v1.0
```

#### 2. Playbook File Not Found
```bash
Error: failed to read playbook file /path/to/playbook.yaml: no such file or directory
```
**Solution**: Verify the file path and permissions:
```bash
# Check file exists
ls -la /path/to/playbook.yaml

# Use absolute path or relative to current directory
kubectl-mtv create hook my-hook \
  --image ansible-runner:latest \
  --playbook @./playbooks/my-playbook.yaml
```

#### 3. Invalid Deadline Value
```bash
Error: invalid hook specification: deadline must be non-negative, got: -100
```
**Solution**: Use positive values for deadline:
```bash
# Correct: positive deadline
kubectl-mtv create hook my-hook \
  --image hook-image:latest \
  --deadline 300

# Remove deadline for no timeout
kubectl-mtv create hook my-hook \
  --image hook-image:latest
```

#### 4. Hook Already Exists
```bash
Error: failed to create hook my-hook: hook 'my-hook' already exists in namespace 'default'
```
**Solution**: Use a different name or delete the existing hook:
```bash
# Delete existing hook
kubectl delete hook my-hook

# Or use a different name
kubectl-mtv create hook my-hook-v2 --image hook-image:latest
```

#### 5. Service Account Not Found
```bash
Error: ServiceAccount "nonexistent-sa" not found
```
**Solution**: Create the service account first or use an existing one:
```bash
# Create service account
kubectl create serviceaccount my-hook-sa

# Then create hook
kubectl-mtv create hook my-hook \
  --image hook-image:latest \
  --service-account my-hook-sa

# Or use default service account
kubectl-mtv create hook my-hook \
  --image hook-image:latest \
  --service-account default
```

### Debugging Hooks

#### View Hook Status

```bash
# Check hook resource status
kubectl get hook my-hook -o yaml

# Look for conditions and status
kubectl describe hook my-hook
```

#### Access Hook Logs

When hooks are executed by migration plans, check pod logs:

```bash
# List pods related to migration
kubectl get pods -l app=migration

# View hook execution logs
kubectl logs <hook-pod-name>
```

#### Validate Playbook Syntax

```bash
# Extract and validate playbook
kubectl get hook my-hook -o jsonpath='{.spec.playbook}' | base64 -d > /tmp/hook-playbook.yaml
ansible-playbook --syntax-check /tmp/hook-playbook.yaml
```

## Security Considerations

### 1. Least Privilege Service Accounts

Create service accounts with minimal required permissions:

```bash
# Create restricted service account
kubectl create serviceaccount hook-sa

# Create role with specific permissions only
kubectl create role hook-role \
  --verb=get,list \
  --resource=configmaps

# Bind role to service account
kubectl create rolebinding hook-binding \
  --role=hook-role \
  --serviceaccount=default:hook-sa
```

### 2. Secure Image Sources

Use trusted container registries and verify image signatures:

```bash
# Use official or trusted images
kubectl-mtv create hook my-hook --image registry.redhat.io/ubi8/ubi:latest

# Or internal registry
kubectl-mtv create hook my-hook --image internal-registry.company.com/hooks/validator:v1.0
```

### 3. Sensitive Data Handling

Use Kubernetes secrets for sensitive data in playbooks:

```yaml
---
- name: Secure Hook with Secrets
  hosts: localhost
  tasks:
    - name: Use secret data
      debug:
        msg: "Database password: {{ lookup('env', 'DB_PASSWORD') }}"
      vars:
        DB_PASSWORD: "{{ lookup('kubernetes.core.k8s', 'v1', 'Secret', 'db-credentials', 'default')['data']['password'] | b64decode }}"
```

This comprehensive guide covers all aspects of creating and managing migration hooks with kubectl-mtv. For more advanced topics and integration patterns, refer to the specific migration planning documentation. 