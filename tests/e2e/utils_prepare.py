import tempfile
import os

def prepare_namespace_for_testing(context):
    """
    Prepare the namespace for testing by creating two NetworkAttachmentDefinitions.
    """
    import yaml
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
