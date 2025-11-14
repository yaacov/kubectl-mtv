---
layout: page
title: "Chapter 14: Migration Hooks"
---

# Chapter 14: Migration Hooks

Migration hooks provide powerful custom automation capabilities that enable running custom code at specific points during the migration process. This chapter covers comprehensive hook development, deployment, and integration with migration workflows.

## Overview: Enabling Custom Automation

### What are Migration Hooks?

Migration hooks are Kubernetes resources that define custom automation to be executed during VM migration:

- **Pre-migration Hooks**: Execute before VM conversion begins
- **Post-migration Hooks**: Execute after VM migration completes
- **Custom Logic**: Handle migration-specific requirements unique to your environment
- **Ansible-Based**: Leverage Ansible playbooks for automation logic

### Hook Execution Model

Hooks run as Kubernetes Jobs in the `konveyor-forklift` namespace:

1. **Job Creation**: Hook container is scheduled as a Kubernetes Job
2. **Context Injection**: Migration context is made available via mounted files
3. **Playbook Execution**: Ansible playbook runs with access to migration data
4. **Result Handling**: Job completion status determines hook success/failure

### Common Use Cases

- **Pre-migration**: Application quiescing, backup creation, dependency checks
- **Post-migration**: Health validation, configuration updates, service restoration
- **Integration**: External system notifications, monitoring updates, compliance logging
- **Troubleshooting**: Debug information collection, state preservation

## Hook Architecture and Context

### Default Hook Image

The default hook runtime is based on Ansible Runner with Kubernetes integration:

- **Default Image**: `quay.io/kubev2v/hook-runner` (verified from command code)
- **Based On**: Ansible Runner with python-openshift and oc binary
- **Capabilities**: Full Ansible automation with Kubernetes API access
- **Extensibility**: Custom images supported for specialized requirements

### Accessing Migration Context

Hooks receive migration context through mounted files:

#### plan.yml
Contains complete migration plan information:
- Plan metadata and configuration
- VM list and specifications  
- Mapping configurations
- Provider references

#### workload.yml
Contains current VM (workload) specific data:
- VM properties and configuration
- IP addresses and network information
- Disk and storage details
- Migration-specific metadata

### Hook Parameters

Hook creation supports comprehensive configuration verified from the command code:

| Parameter | Flag | Description | Default |
|-----------|------|-------------|---------|
| **Image** | `--image`, `-i` | Container image URL | `quay.io/kubev2v/hook-runner` |
| **Playbook** | `--playbook` | Ansible playbook content or @file | None |
| **Service Account** | `--service-account` | Kubernetes ServiceAccount | Default |
| **Deadline** | `--deadline` | Hook timeout in seconds | No timeout |

## How-To: Creating Hooks

### Basic Hook Creation

```bash
# Create hook with inline playbook
kubectl mtv create hook simple-notification \
  --playbook "$(cat << 'EOF'
- hosts: localhost
  tasks:
  - name: Send notification
    debug:
      msg: "Migration hook executed for {{ workload.vm.name }}"
EOF
)"
```

### Hook Creation with File Playbook

```bash
# Create playbook file
cat > database-backup.yml << 'EOF'
- name: Database Backup Hook
  hosts: localhost
  tasks:
  - name: Load migration context
    include_vars:
      file: plan.yml
      name: plan
  
  - name: Load workload context  
    include_vars:
      file: workload.yml
      name: workload
  
  - name: Create database backup
    shell: |
      mysqldump -h {{ workload.vm.ipaddress }} -u backup_user \
      --all-databases > /backup/{{ workload.vm.name }}-$(date +%Y%m%d).sql
    when: workload.vm.name is match('.*database.*')
EOF

# Create hook using file
kubectl mtv create hook database-backup \
  --playbook @database-backup.yml \
  --service-account migration-hooks-sa \
  --deadline 1800
```

### Advanced Hook with Custom Image

```bash
# Create hook with custom image and extended capabilities
kubectl mtv create hook advanced-validation \
  --image registry.company.com/migration/custom-hook:v1.2 \
  --playbook @advanced-validation.yml \
  --service-account admin-hooks-sa \
  --deadline 3600
```

## Detailed Hook Examples

### Example 1: Database Backup Hook (Pre-migration)

#### Create the Playbook

```yaml
# Save as db-backup-hook.yml
- name: Database Backup Pre-Migration Hook
  hosts: localhost
  tasks:
  - name: Load migration plan context
    include_vars:
      file: plan.yml
      name: plan

  - name: Load VM workload context
    include_vars:
      file: workload.yml
      name: workload

  - name: Create backup directory
    file:
      path: /backup/{{ plan.metadata.name }}
      state: directory
      mode: '0755'

  - name: Check if VM is database server
    set_fact:
      is_database: "{{ workload.vm.name | regex_search('(database|db|sql|mysql|postgres)', ignorecase=True) | bool }}"

  - name: Create database backup
    shell: |
      BACKUP_FILE="/backup/{{ plan.metadata.name }}/{{ workload.vm.name }}-backup-$(date +%Y%m%d_%H%M%S).sql"
      mysqldump -h {{ workload.vm.ipaddress }} -u {{ backup_user }} -p{{ backup_password }} \
        --single-transaction --routines --triggers --all-databases > "${BACKUP_FILE}"
      echo "Backup created: ${BACKUP_FILE}"
    environment:
      backup_user: "{{ lookup('kubernetes.core.k8s', api_version='v1', kind='Secret', namespace='migration', resource_name='db-credentials')['data']['username'] | b64decode }}"
      backup_password: "{{ lookup('kubernetes.core.k8s', api_version='v1', kind='Secret', namespace='migration', resource_name='db-credentials')['data']['password'] | b64decode }}"
    when: is_database

  - name: Store backup location in ConfigMap
    kubernetes.core.k8s:
      api_version: v1
      kind: ConfigMap
      name: "{{ plan.metadata.name }}-backups"
      namespace: migration-backups
      definition:
        data:
          "{{ workload.vm.name }}": "/backup/{{ plan.metadata.name }}/{{ workload.vm.name }}-backup-{{ ansible_date_time.epoch }}.sql"
    when: is_database

  - name: Log backup completion
    debug:
      msg: "Database backup completed for {{ workload.vm.name }}"
    when: is_database
```

#### Create the Hook

```bash
# Create database credentials secret
kubectl create secret generic db-credentials \
  --from-literal=username=backup_user \
  --from-literal=password=secure_backup_password \
  -n migration

# Create the hook
kubectl mtv create hook database-backup-pre \
  --playbook @db-backup-hook.yml \
  --service-account migration-admin \
  --deadline 1800
```

### Example 2: Application Health Check Hook (Post-migration)

#### Create the Playbook

```yaml
# Save as health-check-hook.yml
- name: Post-Migration Health Check
  hosts: localhost
  tasks:
  - name: Load migration contexts
    include_vars:
      file: "{{ item }}"
      name: "{{ item | basename | regex_replace('\\.yml$', '') }}"
    loop:
      - plan.yml
      - workload.yml

  - name: Wait for VM to be ready
    kubernetes.core.k8s_info:
      api_version: kubevirt.io/v1
      kind: VirtualMachine
      name: "{{ workload.vm.name }}"
      namespace: "{{ plan.spec.targetNamespace | default('default') }}"
      wait: true
      wait_condition:
        type: Ready
        status: 'True'
      wait_timeout: 600

  - name: Check application endpoints
    uri:
      url: "http://{{ workload.vm.ipaddress }}:{{ item.port }}{{ item.path }}"
      method: GET
      status_code: 200
      timeout: 30
    register: health_checks
    loop:
      - { port: 8080, path: "/health" }
      - { port: 8080, path: "/ready" }
      - { port: 9090, path: "/metrics" }
    ignore_errors: true

  - name: Validate service responses
    assert:
      that:
        - item.status == 200
      fail_msg: "Health check failed for {{ item.url }}"
      success_msg: "Health check passed for {{ item.url }}"
    loop: "{{ health_checks.results }}"
    when: item.status is defined

  - name: Update monitoring configuration
    kubernetes.core.k8s:
      api_version: v1
      kind: ConfigMap
      name: monitoring-targets
      namespace: monitoring
      definition:
        data:
          "{{ workload.vm.name }}": |
            - targets: ['{{ workload.vm.ipaddress }}:9090']
              labels:
                instance: '{{ workload.vm.name }}'
                environment: '{{ plan.metadata.labels.environment | default("production") }}'

  - name: Send notification
    uri:
      url: "{{ notification_webhook }}"
      method: POST
      body_format: json
      body:
        text: "Migration completed for {{ workload.vm.name }} - Health checks passed"
        vm: "{{ workload.vm.name }}"
        plan: "{{ plan.metadata.name }}"
        status: "healthy"
    vars:
      notification_webhook: "{{ lookup('kubernetes.core.k8s', api_version='v1', kind='Secret', namespace='migration', resource_name='notification-config')['data']['webhook_url'] | b64decode }}"
```

#### Create the Hook

```bash
# Create notification webhook secret
kubectl create secret generic notification-config \
  --from-literal=webhook_url=https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
  -n migration

# Create the hook
kubectl mtv create hook health-check-post \
  --playbook @health-check-hook.yml \
  --service-account migration-monitor \
  --deadline 900
```

### Example 3: Shell Script Hook Integration

#### Create Shell Script Hook

```yaml
# Save as shell-script-hook.yml
- name: Shell Script Integration Hook
  hosts: localhost
  tasks:
  - name: Load migration context
    include_vars:
      file: workload.yml
      name: workload

  - name: Execute custom shell script
    script: |
      #!/bin/bash
      
      VM_NAME="{{ workload.vm.name }}"
      VM_IP="{{ workload.vm.ipaddress }}"
      PLAN_NAME="{{ plan.metadata.name | default('unknown') }}"
      
      echo "Processing VM: $VM_NAME (IP: $VM_IP) in plan: $PLAN_NAME"
      
      # Custom application-specific logic
      if [[ "$VM_NAME" == *"web"* ]]; then
        echo "Configuring web server post-migration..."
        # Update load balancer configuration
        curl -X POST "http://load-balancer.internal/api/servers" \
          -H "Content-Type: application/json" \
          -d "{\"name\":\"$VM_NAME\",\"ip\":\"$VM_IP\",\"status\":\"active\"}"
      fi
      
      if [[ "$VM_NAME" == *"database"* ]]; then
        echo "Validating database connectivity..."
        # Test database connection
        timeout 30 bash -c "until nc -z $VM_IP 3306; do sleep 1; done"
        echo "Database is responding on $VM_IP:3306"
      fi
      
      # Update external monitoring
      curl -X POST "http://monitoring.internal/api/targets" \
        -H "Authorization: Bearer $MONITORING_TOKEN" \
        -d "host=$VM_IP&name=$VM_NAME&plan=$PLAN_NAME"
    environment:
      MONITORING_TOKEN: "{{ lookup('kubernetes.core.k8s', api_version='v1', kind='Secret', namespace='migration', resource_name='monitoring-token')['data']['token'] | b64decode }}"
    register: script_result

  - name: Log script execution results
    debug:
      var: script_result.stdout_lines
```

#### Create and Use the Hook

```bash
# Create monitoring token secret
kubectl create secret generic monitoring-token \
  --from-literal=token=your-monitoring-api-token \
  -n migration

# Create the hook
kubectl mtv create hook shell-integration \
  --playbook @shell-script-hook.yml \
  --service-account migration-integration \
  --deadline 600
```

## Adding Hooks via Plan Creation Flags

### Plan-Level Hook Integration

Plan creation supports adding hooks to all VMs in the plan:

```bash
# Add pre-hook to all VMs in the plan
kubectl mtv create plan hooked-migration \
  --source vsphere-prod \
  --pre-hook database-backup-pre \
  --vms "database-01,database-02,app-server-01"

# Add both pre and post hooks
kubectl mtv create plan comprehensive-hooks \
  --source vsphere-prod \
  --pre-hook preparation-hook \
  --post-hook validation-hook \
  --vms "where name ~= '.*prod.*'"

# Combined with other migration settings
kubectl mtv create plan production-with-hooks \
  --source vsphere-prod \
  --target-namespace production \
  --migration-type warm \
  --network-mapping prod-network-map \
  --storage-mapping prod-storage-map \
  --pre-hook backup-and-quiesce \
  --post-hook health-and-notify \
  --vms @production-vms.yaml
```

### Hook Execution Order

When multiple hooks are configured:

1. **Pre-hooks execute**: Before VM conversion begins
2. **Migration proceeds**: VM conversion and data transfer
3. **Post-hooks execute**: After VM migration completes

## Managing Hooks via PlanVM Configuration

### Per-VM Hook Configuration

Individual VMs can have specific hooks using the PlanVMS format:

```yaml
# vm-specific-hooks.yaml
- name: database-primary
  targetName: db-prod-01
  hooks:
  - step: PreHook
    hook:
      name: database-backup-pre
      namespace: migration-hooks
  - step: PostHook
    hook:
      name: database-health-check
      namespace: migration-hooks

- name: web-server-01
  targetName: web-prod-01
  hooks:
  - step: PreHook
    hook:
      name: web-drain-connections
      namespace: migration-hooks
  - step: PostHook
    hook:
      name: web-health-check
      namespace: migration-hooks

- name: cache-server-01
  targetName: cache-prod-01
  hooks:
  - step: PostHook
    hook:
      name: cache-warmup
      namespace: migration-hooks
```

### Using VM-Specific Hooks

```bash
# Create plan with VM-specific hooks
kubectl mtv create plan vm-specific-hooks \
  --source vsphere-prod \
  --vms @vm-specific-hooks.yaml \
  --network-mapping prod-network-map \
  --storage-mapping prod-storage-map
```

### Hook Management via Plan Patching

Hooks can be added or modified after plan creation:

```bash
# Add hook to specific VM in existing plan (requires plan patching - see Chapter 15)
kubectl patch plan existing-plan --type='merge' -p='
spec:
  vms:
  - name: additional-vm
    hooks:
    - step: PreHook
      hook:
        name: new-preparation-hook
        namespace: migration-hooks'
```

## Advanced Hook Development

### Custom Hook Images

#### Building Custom Hook Image

```dockerfile
# Custom hook image with additional tools
FROM quay.io/kubev2v/hook-runner:latest

# Install additional packages
USER root
RUN dnf install -y postgresql mysql jq curl wget && \
    dnf clean all

# Install custom Python packages
RUN pip3 install psycopg2-binary pymongo redis

# Add custom scripts
COPY scripts/ /usr/local/bin/
RUN chmod +x /usr/local/bin/*

# Return to ansible user
USER ansible
```

#### Using Custom Image

```bash
# Build and push custom image
docker build -t registry.company.com/migration/custom-hook:v1.0 .
docker push registry.company.com/migration/custom-hook:v1.0

# Create hook with custom image
kubectl mtv create hook custom-database-hook \
  --image registry.company.com/migration/custom-hook:v1.0 \
  --playbook @database-migration-hook.yml \
  --service-account database-migration-sa
```

### Hook Development Best Practices

#### Error Handling and Logging

```yaml
# Error handling in hook playbooks
- name: Robust Hook with Error Handling
  hosts: localhost
  tasks:
  - name: Set failure flag
    set_fact:
      hook_failed: false

  - name: Critical operation with error handling
    block:
      - name: Perform critical task
        # Task that might fail
        shell: risky_command_here
        register: result
        
    rescue:
      - name: Handle failure
        set_fact:
          hook_failed: true
          
      - name: Log failure details
        debug:
          msg: "Hook failed: {{ ansible_failed_result.msg }}"
          
      - name: Send failure notification
        # Notification logic here
        
    always:
      - name: Cleanup operations
        # Cleanup logic here

  - name: Fail hook if critical operations failed
    fail:
      msg: "Hook execution failed"
    when: hook_failed
```

#### Secret and ConfigMap Access

```yaml
# Secure credential access in hooks
- name: Secure Credential Management
  hosts: localhost
  tasks:
  - name: Load database credentials
    kubernetes.core.k8s_info:
      api_version: v1
      kind: Secret
      name: database-credentials
      namespace: migration-secrets
    register: db_creds

  - name: Use credentials securely
    # Use credentials from db_creds.resources[0].data
    no_log: true  # Don't log sensitive operations
```

#### Timeout and Deadline Management

```yaml
# Timeout management in hook operations
- name: Hook with Timeout Management
  hosts: localhost
  tasks:
  - name: Operation with timeout
    async: 600  # 10 minute timeout
    poll: 10    # Check every 10 seconds
    # Long running operation here
    
  - name: Wait for background task
    async_status:
      jid: "{{ operation_result.ansible_job_id }}"
    register: job_result
    until: job_result.finished
    retries: 60
    delay: 10
```

## Hook Integration Scenarios

### Scenario 1: Enterprise Database Migration

```yaml
# Complete database migration hook workflow
- name: Enterprise Database Migration Hooks
  hosts: localhost
  vars:
    notification_webhook: "{{ lookup('env', 'NOTIFICATION_WEBHOOK') }}"
    
  tasks:
  # Pre-migration tasks
  - name: Database pre-migration validation
    block:
      - name: Check database connectivity
        # Validate source database accessibility
        
      - name: Create backup
        # Full database backup
        
      - name: Quiesce applications
        # Stop application connections
        
      - name: Validate backup integrity
        # Verify backup completed successfully

  # Post-migration tasks  
  - name: Database post-migration validation
    block:
      - name: Validate target database startup
        # Ensure database started correctly
        
      - name: Run data integrity checks
        # Verify data migration completeness
        
      - name: Update DNS records
        # Point applications to new database
        
      - name: Resume application connections
        # Allow applications to reconnect
        
      - name: Send completion notification
        # Notify operations team
```

### Scenario 2: Multi-Tier Application Migration

```yaml
# Coordinated multi-tier application migration
- name: Multi-Tier Application Migration
  hosts: localhost
  tasks:
  - name: Load balancer update
    when: workload.vm.name is match('.*web.*')
    # Remove from load balancer (pre) / Add to load balancer (post)
    
  - name: Session store migration  
    when: workload.vm.name is match('.*cache.*')
    # Handle session data migration
    
  - name: Database coordination
    when: workload.vm.name is match('.*database.*')
    # Database-specific migration logic
    
  - name: Service discovery update
    # Update service registry entries
```

### Scenario 3: Compliance and Auditing

```yaml
# Compliance and audit trail hooks
- name: Compliance Migration Hook
  hosts: localhost
  tasks:
  - name: Log migration start
    # Create audit trail entry
    
  - name: Validate security compliance
    # Check security configurations
    
  - name: Document configuration changes
    # Record all migration changes
    
  - name: Generate compliance report
    # Create detailed compliance documentation
    
  - name: Store audit evidence
    # Archive all compliance artifacts
```

## Hook Monitoring and Troubleshooting

### Hook Execution Monitoring

```bash
# Monitor hook job execution
kubectl get jobs -n konveyor-forklift | grep hook

# Check hook pod logs
kubectl logs -n konveyor-forklift job/hook-job-name

# Monitor hook completion
kubectl get jobs -n konveyor-forklift -w
```

### Hook Debugging

```bash
# Describe failed hook job
kubectl describe job hook-job-name -n konveyor-forklift

# Get hook pod details
kubectl get pods -n konveyor-forklift -l job-name=hook-job-name

# Access hook execution environment
kubectl exec -it hook-pod-name -n konveyor-forklift -- /bin/bash

# Check hook context files
kubectl exec hook-pod-name -n konveyor-forklift -- cat /tmp/hook/plan.yml
kubectl exec hook-pod-name -n konveyor-forklift -- cat /tmp/hook/workload.yml
```

### Common Hook Issues

#### Timeout Issues

```bash
# Check hook deadline configuration
kubectl describe hook hook-name

# Monitor long-running hooks
kubectl logs -f job/hook-job-name -n konveyor-forklift
```

#### Permission Issues

```bash
# Verify ServiceAccount permissions
kubectl describe serviceaccount hook-service-account -n konveyor-forklift

# Check RBAC configurations
kubectl auth can-i create configmaps --as=system:serviceaccount:konveyor-forklift:hook-service-account
```

#### Playbook Errors

```bash
# Validate playbook syntax
ansible-playbook --syntax-check playbook.yml

# Test playbook locally (with mock context)
ansible-playbook -i localhost, playbook.yml
```

## Hook Security and Best Practices

### Security Considerations

1. **Least Privilege**: Use minimal required ServiceAccount permissions
2. **Secret Management**: Store sensitive data in Kubernetes Secrets
3. **Network Isolation**: Limit hook network access when possible
4. **Image Security**: Use trusted, regularly updated hook images

### Operational Best Practices

1. **Testing**: Thoroughly test hooks in non-production environments
2. **Idempotency**: Design hooks to be safely re-runnable
3. **Monitoring**: Implement comprehensive hook execution monitoring
4. **Documentation**: Maintain clear hook documentation and runbooks

### Performance Optimization

1. **Efficient Playbooks**: Minimize unnecessary operations in playbooks
2. **Parallel Execution**: Use Ansible parallel features where appropriate
3. **Resource Limits**: Set appropriate resource limits on hook jobs
4. **Cleanup**: Implement proper cleanup for temporary resources

## Next Steps

After mastering migration hooks:

1. **Advanced Plan Management**: Learn dynamic plan modification in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)
2. **Execute Migrations**: Manage complete migration lifecycle in [Chapter 16: Plan Lifecycle Execution](16-plan-lifecycle-execution.md)
3. **Troubleshooting**: Master debugging techniques in [Chapter 17: Debugging and Troubleshooting](17-debugging-and-troubleshooting.md)
4. **Best Practices**: Learn operational excellence in [Chapter 18: Best Practices and Security](18-best-practices-and-security.md)

---

*Previous: [Chapter 13: Migration Process Optimization](13-migration-process-optimization.md)*  
*Next: [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)*
