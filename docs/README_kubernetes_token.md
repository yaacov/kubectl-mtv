# Service Account Admin Access & Token Retrieval

This guide shows how to create a **ServiceAccount** that is an **admin in a single namespace (project)** and how to obtain an authentication token for it.

> **Prerequisites**
>
> * `kubectl` or `oc` configured to reach the cluster with permissions to create namespaces/projects and RBAC objects.
> * Kubernetes ≥ 1.24 or OpenShift ≥ 4.11 recommended (commands use the modern _TokenRequest_ API).

---

## 1. Kubernetes (Vanilla)

Create a ServiceAccount and bind it to the built‑in `admin` ClusterRole within a namespace.

| Step | Command | Notes |
|------|---------|-------|
| 1. Create or switch to a namespace | `kubectl create namespace demo-ns` | Skip if it already exists. |
| 2. Create the service account | `kubectl create serviceaccount demo-admin -n demo-ns` | |
| 3. Bind _admin_ rights inside the namespace | `kubectl create rolebinding demo-admin-binding \`<br>`--clusterrole=admin \`<br>`--serviceaccount=demo-ns:demo-admin \`<br>`--namespace demo-ns` | Uses the cluster‑wide **admin** role, but the binding scopes it to `demo-ns`. |
| 4. Generate a short‑lived token (recommended) | `kubectl -n demo-ns create token demo-admin --duration=24h` | Prints a JWT usable immediately. Requires K8s ≥ 1.24. |

---

## 2. OpenShift (Container Platform)

Use `oc` commands to provision a ServiceAccount and grant namespace‑scoped admin rights.

| Step | Command | Notes |
|------|---------|-------|
| 1. Create a project (namespace) | `oc new-project demo-ns` | Projects add default quotas/SCCs. |
| 2. Create the service account | `oc create sa demo-admin -n demo-ns` | |
| 3. Grant namespace admin | `oc adm policy add-role-to-user admin -z demo-admin -n demo-ns` | `-z` targets a service account. |
| 4. Obtain a token | `oc -n demo-ns create token demo-admin --duration=24h` | Works on OCP ≥ 4.11. |

---

## 3. Verification

Confirm the ServiceAccount can list resources only in its own namespace and is denied elsewhere.

```bash
# Should list resources in the namespace
kubectl --context=demo-admin get all -n demo-ns

# Should be forbidden in other namespaces
kubectl --context=demo-admin get pods -A
```

---

## 4. Cleanup

Revoke roles and delete the ServiceAccount and namespace/project.

```bash
# Delete RoleBinding / policy
kubectl delete rolebinding demo-admin-binding -n demo-ns
oc adm policy remove-role-from-user admin -z demo-admin -n demo-ns

# Delete SA and namespace
kubectl delete sa demo-admin -n demo-ns && kubectl delete ns demo-ns
```

---

### Long‑term Token Generation

Create a Secret to hold a reusable ServiceAccount token.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demo-admin-token
  annotations:
    kubernetes.io/service-account.name: demo-admin
  namespace: demo-ns
type: kubernetes.io/service-account-token
```

```bash
kubectl apply -f secret.yaml
kubectl -n demo-ns get secret demo-admin-token -o go-template='{{ .data.token | base64decode }}'
```

The control plane populates the token field shortly after the Secret is created.
