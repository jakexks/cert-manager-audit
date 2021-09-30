ARG KUBERNETES_VERSION=v1.22.1
FROM kindest/node:${KUBERNETES_VERSION}
COPY audit-policy.yaml /config/audit-policy.yaml
COPY audit-webhook-config.yaml /config/audit-webhook-config.yaml
