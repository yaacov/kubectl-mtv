import tempfile
import os
import yaml
from .utils import (
    generate_provider_name,
    get_or_create_provider,
)
from .test_constants import (
    NETWORK_ATTACHMENT_DEFINITIONS,
    OPENSHIFT_TEST_VMS,
)


def prepare_namespace_for_testing(context):
    """
    Prepare the namespace for testing by creating test artifacts.
    """

    # Create OpenShift target provider for plan tests
    target_provider_name = generate_provider_name(
        "openshift", "localhost", skip_tls=True
    )
    create_cmd = f"create provider {target_provider_name} --type openshift"

    # Create provider if it doesn't already exist
    get_or_create_provider(context, target_provider_name, create_cmd)

    # Create two NetworkAttachmentDefinitions in the test namespace
    for nad_name in NETWORK_ATTACHMENT_DEFINITIONS:
        nad_obj = {
            "apiVersion": "k8s.cni.cncf.io/v1",
            "kind": "NetworkAttachmentDefinition",
            "metadata": {"name": nad_name, "namespace": context.namespace},
            "spec": {
                "config": '{"cniVersion": "0.3.1", "type": "bridge", "bridge": "br0", "ipam": {"type": "host-local", "subnet": "10.10.0.0/16"}}'
            },
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
    for idx, vm_name in enumerate(OPENSHIFT_TEST_VMS):
        vm_obj = {
            "apiVersion": "kubevirt.io/v1",
            "kind": "VirtualMachine",
            "metadata": {"name": vm_name, "namespace": context.namespace},
            "spec": {
                "dataVolumeTemplates": [
                    {
                        "metadata": {"name": f"{vm_name}-volume"},
                        "spec": {
                            "sourceRef": {
                                "kind": "DataSource",
                                "name": "centos-stream9",
                                "namespace": "openshift-virtualization-os-images",
                            },
                            "storage": {"resources": {"requests": {"storage": "30Gi"}}},
                        },
                    }
                ],
                "instancetype": {
                    "kind": "virtualmachineclusterinstancetype",
                    "name": "u1.medium",
                },
                "preference": {
                    "kind": "virtualmachineclusterpreference",
                    "name": "centos.stream9",
                },
                "runStrategy": "Halted",
                "template": {
                    "metadata": {
                        "creationTimestamp": None,
                        "labels": {"network.kubevirt.io/headlessService": "headless"},
                    },
                    "spec": {
                        "architecture": "amd64",
                        "domain": {
                            "devices": {
                                "autoattachPodInterface": False,
                                "interfaces": [
                                    {"masquerade": {}, "name": "default"},
                                    {
                                        "name": NETWORK_ATTACHMENT_DEFINITIONS[0],
                                        "bridge": {},
                                    },
                                    {
                                        "name": NETWORK_ATTACHMENT_DEFINITIONS[1],
                                        "bridge": {},
                                    },
                                ],
                            },
                            "machine": {"type": "pc-q35-rhel9.6.0"},
                            "resources": {},
                        },
                        "networks": [
                            {"name": "default", "pod": {}},
                            {
                                "name": NETWORK_ATTACHMENT_DEFINITIONS[0],
                                "multus": {
                                    "networkName": f"{context.namespace}/{NETWORK_ATTACHMENT_DEFINITIONS[0]}"
                                },
                            },
                            {
                                "name": NETWORK_ATTACHMENT_DEFINITIONS[1],
                                "multus": {
                                    "networkName": f"{context.namespace}/{NETWORK_ATTACHMENT_DEFINITIONS[1]}"
                                },
                            },
                        ],
                        "subdomain": "headless",
                        "volumes": [
                            {
                                "dataVolume": {"name": f"{vm_name}-volume"},
                                "name": "rootdisk",
                            },
                            {
                                "cloudInitNoCloud": {
                                    "userData": """#cloud-config\nchpasswd:\n  expire: false\n  password: i90d-diu0-m9ci\n  user: centos\n"""
                                },
                                "name": "cloudinitdisk",
                            },
                        ],
                    },
                },
            },
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
