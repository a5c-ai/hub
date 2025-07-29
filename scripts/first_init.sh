# Make sure you're logged in to Azure
# az login --tenant $AZURE_TENANT_ID

# Create a resource group for Terraform state (separate from your main infrastructure)
RG_NAME="a5ctf-rg"
STORAGE_ACCOUNT_NAME="a5ctfstorage"
CONTAINER_NAME="tf-main"

az group create --name $RG_NAME --location "East US" --subscription $AZURE_SUBSCRIPTION_ID

# Create a storage account for Terraform state
az storage account create \
  --name $STORAGE_ACCOUNT_NAME \
  --resource-group $RG_NAME \
  --location "East US" \
  --sku Standard_LRS \
  --subscription $AZURE_SUBSCRIPTION_ID

# Create a container for the state files
az storage container create \
  --name $CONTAINER_NAME \
  --account-name $STORAGE_ACCOUNT_NAME \
  --subscription $AZURE_SUBSCRIPTION_ID

# Print the storage account name for later use
echo "Storage Account Name: $STORAGE_ACCOUNT_NAME"