import tempfile
import os
import yaml
from .utils import wait_for_provider_ready

def prepare_namespace_for_testing(context):
    """
    Prepare the namespace for testing by creating test artifacts.
    """

    # Create OpenShift target provider for plan tests
    target_provider_name = "test-openshift-target"
    
    # Create a simple OpenShift provider using current cluster context
    create_cmd = f"create provider {target_provider_name} --type openshift"
    
    # Create provider
    result = context.run_mtv_command(create_cmd)
    if result.returncode == 0:
        # Track for cleanup
        context.track_resource("provider", target_provider_name)
        
        # Wait for provider to be ready
        wait_for_provider_ready(context, target_provider_name)

    # Create two NetworkAttachmentDefinitions in the test namespace
    nad_names = ["test-nad-1", "test-nad-2"]
    for nad_name in nad_names:
        nad_obj = {
            "apiVersion": "k8s.cni.cncf.io/v1",
            "kind": "NetworkAttachmentDefinition",
            "metadata": {
                "name": nad_name,
                "namespace": context.namespace
            },
            "spec": {
                "config": '{"cniVersion": "0.3.1", "type": "bridge", "bridge": "br0", "ipam": {"type": "host-local", "subnet": "10.10.0.0/16"}}'
            }
        }
        manifest_yaml = yaml.dump(nad_obj)
        with tempfile.NamedTemporaryFile(mode="w", delete=False) as f:
            f.write(manifest_yaml)
            temp_path = f.name
        try:
            context.run_kubectl_command(f"apply -f {temp_path}")
            context.track_resource("network-attachment-definition", nad_name)
        finally:
            os.unlink(temp_path)
    
    # Create two VirtualMachines in the test namespace
    vm_names = ["test-vm-1", "test-vm-2"]
    for idx, vm_name in enumerate(vm_names):
        vm_obj = {
            "apiVersion": "kubevirt.io/v1",
            "kind": "VirtualMachine",
            "metadata": {
                "name": vm_name,
                "namespace": context.namespace
            },
            "spec": {
                "dataVolumeTemplates": [
                    {
                        "metadata": {"name": f"{vm_name}-volume"},
                        "spec": {
                            "sourceRef": {
                                "kind": "DataSource",
                                "name": "centos-stream9",
                                "namespace": "openshift-virtualization-os-images"
                            },
                            "storage": {
                                "resources": {"requests": {"storage": "30Gi"}}
                            }
                        }
                    }
                ],
                "instancetype": {
                    "kind": "virtualmachineclusterinstancetype",
                    "name": "u1.medium"
                },
                "preference": {
                    "kind": "virtualmachineclusterpreference",
                    "name": "centos.stream9"
                },
                "runStrategy": "Halted",
                "template": {
                    "metadata": {
                        "creationTimestamp": None,
                        "labels": {"network.kubevirt.io/headlessService": "headless"}
                    },
                    "spec": {
                        "architecture": "amd64",
                        "domain": {
                            "devices": {
                                "autoattachPodInterface": False,
                                "interfaces": [
                                    {
                                        "masquerade": {},
                                        "name": "default"
                                    },
                                    {
                                        "name": "test-nad-1",
                                        "bridge": {}
                                    },
                                    {
                                        "name": "test-nad-2",
                                        "bridge": {}
                                    }
                                ]
                            },
                            "machine": {"type": "pc-q35-rhel9.6.0"},
                            "resources": {}
                        },
                        "networks": [
                            {"name": "default", "pod": {}},
                            {"name": "test-nad-1", "multus": {"networkName": f"{context.namespace}/test-nad-1"}},
                            {"name": "test-nad-2", "multus": {"networkName": f"{context.namespace}/test-nad-2"}}
                        ],
                        "subdomain": "headless",
                        "volumes": [
                            {
                                "dataVolume": {"name": f"{vm_name}-volume"},
                                "name": "rootdisk"
                            },
                            {
                                "cloudInitNoCloud": {
                                    "userData": """#cloud-config\nchpasswd:\n  expire: false\n  password: i90d-diu0-m9ci\n  user: centos\n"""
                                },
                                "name": "cloudinitdisk"
                            }
                        ]
                    }
                }
            }
        }
        vm_manifest_yaml = yaml.dump(vm_obj)
        with tempfile.NamedTemporaryFile(mode="w", delete=False) as f:
            f.write(vm_manifest_yaml)
            temp_path = f.name
        try:
            context.run_kubectl_command(f"apply -f {temp_path}")
            context.track_resource("virtual-machine", vm_name)
        finally:
            os.unlink(temp_path)
