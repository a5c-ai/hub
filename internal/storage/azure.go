package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	azblob "github.com/Azure/azure-storage-blob-go/azblob"
)

// AzureBackend implements the Backend interface using Azure Blob Storage
// AzureBackend implements the Backend interface using Azure Blob Storage
type AzureBackend struct {
	config       AzureConfig
	containerURL azblob.ContainerURL
}

// NewAzureBackend creates a new Azure Blob Storage backend
func NewAzureBackend(config AzureConfig) (*AzureBackend, error) {
	if config.AccountName == "" {
		return nil, fmt.Errorf("azure account name is required")
	}
	if config.AccountKey == "" {
		return nil, fmt.Errorf("azure account key is required")
	}
	if config.ContainerName == "" {
		return nil, fmt.Errorf("azure container name is required")
	}

	endpoint := config.EndpointURL
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", config.AccountName)
	}

	credential, err := azblob.NewSharedKeyCredential(config.AccountName, config.AccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure credential: %w", err)
	}

	// Construct container URL
	containerURL, err := url.Parse(fmt.Sprintf("%s/%s", endpoint, config.ContainerName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse container URL: %w", err)
	}
	container := azblob.NewContainerURL(*containerURL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	// Create container if not exists
	ctx := context.Background()
	_, err = container.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		if serr, ok := err.(azblob.StorageError); !ok || serr.ServiceCode() != azblob.ServiceCodeContainerAlreadyExists {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}
	}

	return &AzureBackend{
		config:       config,
		containerURL: container,
	}, nil
}

// Upload uploads a file to Azure Blob Storage
func (a *AzureBackend) Upload(ctx context.Context, path string, reader io.Reader, size int64) error {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	_, err := azblob.UploadStreamToBlockBlob(ctx, reader, blobURL,
		azblob.UploadStreamToBlockBlobOptions{BufferSize: 4 * 1024 * 1024, MaxBuffers: 16})
	if err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}
	return nil
}

// Download downloads a file from Azure Blob Storage
func (a *AzureBackend) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	resp, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}
	return resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 3}), nil
}

// Delete deletes a file from Azure Blob Storage
func (a *AzureBackend) Delete(ctx context.Context, path string) error {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}
	return nil
}

// Exists checks if a file exists in Azure Blob Storage
func (a *AzureBackend) Exists(ctx context.Context, path string) (bool, error) {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	_, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok && serr.ServiceCode() == azblob.ServiceCodeBlobNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get blob properties: %w", err)
	}
	return true, nil
}

// GetSize returns the size of a file in bytes
func (a *AzureBackend) GetSize(ctx context.Context, path string) (int64, error) {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	props, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get blob properties: %w", err)
	}
	return props.ContentLength(), nil
}

// GetLastModified returns the last modified time of a file
func (a *AzureBackend) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	blobURL := a.containerURL.NewBlockBlobURL(path)
	props, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get blob properties: %w", err)
	}
	return props.LastModified(), nil
}

// List lists files with the given prefix
func (a *AzureBackend) List(ctx context.Context, prefix string) ([]string, error) {
	marker := azblob.Marker{}
	var blobs []string
	for marker.NotDone() {
		list, err := a.containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{Prefix: prefix})
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}
		for _, v := range list.Segment.BlobItems {
			blobs = append(blobs, v.Name)
		}
		marker = list.NextMarker
	}
	return blobs, nil
}

// GetURL returns a presigned URL for downloading
func (a *AzureBackend) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	credential, err := azblob.NewSharedKeyCredential(a.config.AccountName, a.config.AccountKey)
	if err != nil {
		return "", fmt.Errorf("failed to create azure credential: %w", err)
	}
	sasValues := azblob.BlobSASSignatureValues{
		ContainerName: a.config.ContainerName,
		BlobName:      path,
		Protocol:      azblob.SASProtocolHTTPS,
		StartTime:     time.Now(),
		ExpiryTime:    time.Now().Add(expiry),
		Permissions:   azblob.BlobSASPermissions{Read: true}.String(),
	}
	qs, err := sasValues.NewSASQueryParameters(credential)
	if err != nil {
		return "", fmt.Errorf("failed to sign SAS token: %w", err)
	}
	blobURL := a.containerURL.NewBlockBlobURL(path)
	urlVal := blobURL.URL()
	return fmt.Sprintf("%s?%s", urlVal.String(), qs.Encode()), nil
}
