#!/usr/bin/env bash
set -euo pipefail

LOCATION="${AZURE_LOCATION:-eastus}"
RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:-pulseops-rg}"
ACR_NAME="${AZURE_ACR_NAME:-pulseopsregistry}"
KEY_VAULT_NAME="${AZURE_KEY_VAULT_NAME:-pulseops-kv}"
CONTAINER_ENV_NAME="${AZURE_CONTAINER_ENV_NAME:-pulseops-env}"
CONTAINER_APP_NAME="${AZURE_CONTAINER_APP_NAME:-pulseops-api}"
STATIC_WEB_APP_NAME="${AZURE_STATIC_WEB_APP_NAME:-pulseops-frontend}"
IMAGE="${ACR_NAME}.azurecr.io/pulseops-api:latest"
PLACEHOLDER_IMAGE="mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"

required_vars=(
  AZURE_SUBSCRIPTION_ID
  GOOGLE_CLIENT_SECRET
  MONGODB_ATLAS_URI
)

for var in "${required_vars[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "Missing required environment variable: ${var}" >&2
    exit 1
  fi
done

step() {
  echo
  echo "==> $*"
}

secret_uri() {
  local secret_name="$1"
  az keyvault secret show \
    --vault-name "${KEY_VAULT_NAME}" \
    --name "${secret_name}" \
    --query id \
    --output tsv
}

set_secret_if_missing() {
  local name="$1"
  local value="$2"
  if az keyvault secret show --vault-name "${KEY_VAULT_NAME}" --name "${name}" >/dev/null 2>&1; then
    echo "Secret ${name} already exists; leaving current value in place."
  else
    az keyvault secret set --vault-name "${KEY_VAULT_NAME}" --name "${name}" --value "${value}" >/dev/null
  fi
}

ensure_role_assignment() {
  local assignee="$1"
  local role="$2"
  local scope="$3"
  local existing
  existing="$(az role assignment list --assignee "${assignee}" --role "${role}" --scope "${scope}" --query "length(@)" --output tsv 2>/dev/null || echo 0)"
  if [[ "${existing}" == "0" ]]; then
    az role assignment create --assignee "${assignee}" --role "${role}" --scope "${scope}" >/dev/null
  else
    echo "Role ${role} is already assigned on ${scope}."
  fi
}

step "Selecting Azure subscription ${AZURE_SUBSCRIPTION_ID}"
az account set --subscription "${AZURE_SUBSCRIPTION_ID}"

step "Creating resource group ${RESOURCE_GROUP} in ${LOCATION}"
az group create --name "${RESOURCE_GROUP}" --location "${LOCATION}" >/dev/null

step "Creating Azure Container Registry ${ACR_NAME}"
if ! az acr show --name "${ACR_NAME}" --resource-group "${RESOURCE_GROUP}" >/dev/null 2>&1; then
  az acr create --resource-group "${RESOURCE_GROUP}" --name "${ACR_NAME}" --sku Basic --admin-enabled false >/dev/null
else
  echo "ACR ${ACR_NAME} already exists."
fi
ACR_ID="$(az acr show --name "${ACR_NAME}" --resource-group "${RESOURCE_GROUP}" --query id --output tsv)"

step "Creating Azure Key Vault ${KEY_VAULT_NAME}"
if ! az keyvault show --name "${KEY_VAULT_NAME}" --resource-group "${RESOURCE_GROUP}" >/dev/null 2>&1; then
  az keyvault create --resource-group "${RESOURCE_GROUP}" --name "${KEY_VAULT_NAME}" --location "${LOCATION}" --enable-rbac-authorization true >/dev/null
else
  echo "Key Vault ${KEY_VAULT_NAME} already exists."
fi
KEY_VAULT_ID="$(az keyvault show --name "${KEY_VAULT_NAME}" --resource-group "${RESOURCE_GROUP}" --query id --output tsv)"
CURRENT_USER_ID="$(az ad signed-in-user show --query id --output tsv 2>/dev/null || true)"
if [[ -n "${CURRENT_USER_ID}" ]]; then
  ensure_role_assignment "${CURRENT_USER_ID}" "Key Vault Secrets Officer" "${KEY_VAULT_ID}"
fi

step "Storing required secrets in Key Vault"
JWT_SECRET="$(openssl rand -hex 16)"
set_secret_if_missing "JWT-SECRET" "${JWT_SECRET}"
az keyvault secret set --vault-name "${KEY_VAULT_NAME}" --name "GOOGLE-CLIENT-SECRET" --value "${GOOGLE_CLIENT_SECRET}" >/dev/null
az keyvault secret set --vault-name "${KEY_VAULT_NAME}" --name "MONGODB-URI" --value "${MONGODB_ATLAS_URI}" >/dev/null

step "Creating Container Apps environment ${CONTAINER_ENV_NAME}"
if ! az containerapp env show --name "${CONTAINER_ENV_NAME}" --resource-group "${RESOURCE_GROUP}" >/dev/null 2>&1; then
  az containerapp env create --name "${CONTAINER_ENV_NAME}" --resource-group "${RESOURCE_GROUP}" --location "${LOCATION}" >/dev/null
else
  echo "Container Apps environment ${CONTAINER_ENV_NAME} already exists."
fi

step "Creating or updating Container App ${CONTAINER_APP_NAME}"
if ! az containerapp show --name "${CONTAINER_APP_NAME}" --resource-group "${RESOURCE_GROUP}" >/dev/null 2>&1; then
  az containerapp create \
    --name "${CONTAINER_APP_NAME}" \
    --resource-group "${RESOURCE_GROUP}" \
    --environment "${CONTAINER_ENV_NAME}" \
    --image "${PLACEHOLDER_IMAGE}" \
    --target-port 8080 \
    --ingress external \
    --min-replicas 1 \
    --max-replicas 3 \
    --system-assigned >/dev/null
else
  az containerapp identity assign --name "${CONTAINER_APP_NAME}" --resource-group "${RESOURCE_GROUP}" --system-assigned >/dev/null
  az containerapp update --name "${CONTAINER_APP_NAME}" --resource-group "${RESOURCE_GROUP}" --min-replicas 1 --max-replicas 3 >/dev/null
fi

PRINCIPAL_ID="$(az containerapp show --name "${CONTAINER_APP_NAME}" --resource-group "${RESOURCE_GROUP}" --query identity.principalId --output tsv)"
ensure_role_assignment "${PRINCIPAL_ID}" "Key Vault Secrets User" "${KEY_VAULT_ID}"
ensure_role_assignment "${PRINCIPAL_ID}" "AcrPull" "${ACR_ID}"

step "Configuring ACR access and Key Vault-backed secrets for ${CONTAINER_APP_NAME}"
az containerapp registry set \
  --name "${CONTAINER_APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --server "${ACR_NAME}.azurecr.io" \
  --identity system >/dev/null

JWT_SECRET_URI="$(secret_uri "JWT-SECRET")"
GOOGLE_SECRET_URI="$(secret_uri "GOOGLE-CLIENT-SECRET")"
MONGODB_URI="$(secret_uri "MONGODB-URI")"

az containerapp secret set \
  --name "${CONTAINER_APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --secrets \
    "jwt-secret=keyvaultref:${JWT_SECRET_URI},identityref:system" \
    "google-client-secret=keyvaultref:${GOOGLE_SECRET_URI},identityref:system" \
    "mongodb-uri=keyvaultref:${MONGODB_URI},identityref:system" >/dev/null

az containerapp update \
  --name "${CONTAINER_APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --set-env-vars \
    "JWT_SECRET=secretref:jwt-secret" \
    "GOOGLE_CLIENT_SECRET=secretref:google-client-secret" \
    "MONGODB_URI=secretref:mongodb-uri" \
    "PORT=8080" >/dev/null

step "Enabling sticky sessions for WebSocket subscriptions"
az containerapp ingress sticky-sessions set \
  --name "${CONTAINER_APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --affinity sticky >/dev/null

step "Creating Azure Static Web App ${STATIC_WEB_APP_NAME}"
if ! az staticwebapp show --name "${STATIC_WEB_APP_NAME}" --resource-group "${RESOURCE_GROUP}" >/dev/null 2>&1; then
  az staticwebapp create \
    --name "${STATIC_WEB_APP_NAME}" \
    --resource-group "${RESOURCE_GROUP}" \
    --location "${LOCATION}" \
    --sku Free >/dev/null
else
  echo "Static Web App ${STATIC_WEB_APP_NAME} already exists."
fi

API_FQDN="$(az containerapp show --name "${CONTAINER_APP_NAME}" --resource-group "${RESOURCE_GROUP}" --query properties.configuration.ingress.fqdn --output tsv)"
SWA_URL="$(az staticwebapp show --name "${STATIC_WEB_APP_NAME}" --resource-group "${RESOURCE_GROUP}" --query defaultHostname --output tsv)"

echo
echo "Deployment targets are ready:"
echo "API URL: https://${API_FQDN}"
echo "Static Web App URL: https://${SWA_URL}"
echo
echo "Use this image for first deployment: ${IMAGE}"
