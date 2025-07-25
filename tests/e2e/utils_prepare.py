import tempfile
import os

def prepare_namespace_for_testing(context):
    """
    Prepare the namespace for testing by creating two NetworkAttachmentDefinitions.
    """
    nad1 = "test-nad-1"
    nad2 = "test-nad-2"
    nad_manifest = '''
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: {nad_name}
  namespace: {namespace}
spec:
  config: '{"cniVersion": "0.3.1", "type": "bridge", "bridge": "br0", "ipam": {"type": "host-local", "subnet": "10.10.0.0/16"}}'
'''
    for nad_name in [nad1, nad2]:
        manifest = nad_manifest.format(nad_name=nad_name, namespace=context.namespace)
        # Write manifest to a temp file
        with tempfile.NamedTemporaryFile(mode="w", delete=False) as f:
            f.write(manifest)
            temp_path = f.name
        try:
            context.run_kubectl_command(f"apply -f {temp_path}")
            context.track_resource("network-attachment-definition", nad_name)
        finally:
            os.unlink(temp_path)
