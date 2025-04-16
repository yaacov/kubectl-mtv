# Editing the VMs List for Migration Plans (`planvms`)

When using `kubectl mtv inventory vms <provider> -o planvms`, you get a list of VMs in a format suitable for use with migration plans. This file can be edited before running a migration plan to customize migration behavior for each VM.

## VM List Format

Each VM entry in the list has the following structure:

```yaml
- id: vm-105715
  name: ronen-oc
  # ...other fields...
  hooks: []
  luks:
    kind: ""
    namespace: ""
    name: ""
    uid: ""
    apiversion: ""
    resourceversion: ""
    fieldpath: ""
  rootDisk: ""
  instanceType: ""
  pvcNameTemplate: ""
  volumeNameTemplate: ""
  networkNameTemplate: ""
  targetName: ""
```

### Field Descriptions

- **id**: (string) The provider-specific VM identifier (required).
- **name**: (string) The VM's name in the source provider (required).
- **hooks**: (array) List of migration hooks to apply to this VM (optional).
- **luks**: (object) Reference to a Secret for disk decryption keys (optional).
- **rootDisk**: (string) The primary disk to boot from (optional).
- **instanceType**: (string) Override the VM's instance type in the target (optional).
- **pvcNameTemplate**: (string) Go template for naming PVCs for this VM's disks (optional).
- **volumeNameTemplate**: (string) Go template for naming volume interfaces (optional).
- **networkNameTemplate**: (string) Go template for naming network interfaces (optional).
- **targetName**: (string) Custom name for the VM in the target cluster (optional).

See the code comments in the source for detailed template variable documentation.

## Editing the List

You can edit the YAML file to:

- Remove VMs you do not want to migrate.
- Add or modify fields for specific VMs to customize migration behavior.
- Set per-VM templates for PVC, volume, or network interface names.
- Specify a custom target name for the migrated VM.
- Attach hooks or disk decryption secrets as needed.

**Example:**

```yaml
- id: vm-105715
  name: ronen-oc
  targetName: "ronen-migrated"
  pvcNameTemplate: "{{.VmName}}-disk-{{.DiskIndex}}"
  hooks:
    - name: my-pre-migration-hook
      namespace: migration-hooks
  luks:
    kind: Secret
    namespace: migration-secrets
    name: luks-key-secret
```

## Using the Edited List

After editing, use the file as input to a migration plan:

```bash
kubectl mtv plan create my-plan --source <provider> --target <target-provider> --vms @vms.yaml
```

Replace `vms.yaml` with your edited file.

## Tips

- Only include VMs you want to migrate.
- Ensure all required fields (`id`, `name`) are present for each VM.
- Use per-VM templates to override plan-level settings as needed.
- For advanced customization, refer to the Go template documentation and the comments in the VM struct.
