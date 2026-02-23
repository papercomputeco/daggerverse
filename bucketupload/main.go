package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"dagger/bucketuploader/internal/dagger"
)

const (
	nightly = "nightly"
	latest  = "latest"
)

// FileMetadata holds optional upload headers for a file.
type FileMetadata struct {
	// Base64-encoded SHA-256 checksum of the file contents.
	// When set, the x-amz-checksum-sha256 header is sent with the upload.
	// +optional
	ChecksumSHA256 string

	// MIME content type for the file (e.g., "application/octet-stream").
	// When set, the Content-Type header is sent with the upload.
	// +optional
	ContentType string
}

// FilePathMetadata pairs a relative file path with its upload metadata.
// Used by directory-based upload methods to apply per-file headers.
type FilePathMetadata struct {
	// Relative path of the file inside the artifacts directory
	// (e.g., "bin/my-binary").
	Path string

	// Upload metadata for this file.
	Meta FileMetadata
}

// Bucketuploader provides bucket upload artifact capabilities.
// It expects an S3-compatible bucket via the AWS CLI.
type Bucketuploader struct {
	// Bucket endpoint URL
	//
	// +private
	Endpoint *dagger.Secret

	// Bucket name
	//
	// +private
	Bucket *dagger.Secret

	// Bucket access key ID
	//
	// +private
	AccessKeyID *dagger.Secret

	// Bucket secret access key
	//
	// +private
	SecretAccessKey *dagger.Secret
}

// New creates a new BucketUpload instance configured with bucket credentials.
func New(
	// Bucket endpoint URL
	endpoint *dagger.Secret,

	// Bucket name
	bucket *dagger.Secret,

	// Bucket access key ID
	accessKeyID *dagger.Secret,

	// Bucket secret access key
	secretAccessKey *dagger.Secret,
) *Bucketuploader {
	return &Bucketuploader{
		Endpoint:        endpoint,
		Bucket:          bucket,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}
}

// metadataIndex maps cleaned relative file paths to their metadata.
type metadataIndex map[string]FileMetadata

// buildMetadataIndex creates a lookup map from a slice of FilePathMetadata,
// normalizing paths by stripping leading "./" and "/" prefixes.
func buildMetadataIndex(metadata []FilePathMetadata) metadataIndex {
	idx := make(metadataIndex, len(metadata))
	for _, m := range metadata {
		key := strings.TrimPrefix(strings.TrimPrefix(m.Path, "./"), "/")
		idx[key] = m.Meta
	}
	return idx
}

// upload syncs a directory to the bucket under the given prefix.
// When metadata is provided, files that have metadata entries are uploaded
// individually with the appropriate headers via "aws s3 cp". Files without
// metadata entries are still synced in bulk via "aws s3 sync".
func (b *Bucketuploader) upload(
	ctx context.Context,
	artifacts *dagger.Directory,
	prefix string,
	metadata []FilePathMetadata,
) error {
	bucketName, err := b.Bucket.Plaintext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bucket name: %w", err)
	}

	endpointURL, err := b.Endpoint.Plaintext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get endpoint: %w", err)
	}

	destination := fmt.Sprintf("s3://%s", path.Join(bucketName, prefix))

	awsCli := dag.Container().
		From("amazon/aws-cli:latest").
		WithSecretVariable("AWS_ACCESS_KEY_ID", b.AccessKeyID).
		WithSecretVariable("AWS_SECRET_ACCESS_KEY", b.SecretAccessKey).
		WithEnvVariable("AWS_DEFAULT_REGION", "auto").
		WithDirectory("/artifacts", artifacts).
		WithWorkdir("/artifacts")

	if len(metadata) == 0 {
		// Fast path: no per-file metadata, use bulk sync.
		_, err = awsCli.
			WithExec([]string{
				"aws", "s3", "sync", ".",
				destination,
				"--endpoint-url", endpointURL,
			}).
			Sync(ctx)
		if err != nil {
			return fmt.Errorf("failed to upload artifacts to %s: %w", destination, err)
		}
		return nil
	}

	// Build a lookup of files that have metadata.
	idx := buildMetadataIndex(metadata)

	// List all files in the artifacts directory.
	entries, err := artifacts.Glob(ctx, "**/*")
	if err != nil {
		return fmt.Errorf("failed to list artifact files: %w", err)
	}

	// Upload each file individually: files with metadata get extra headers,
	// files without metadata are uploaded with a plain cp.
	for _, entry := range entries {
		fileDest := fmt.Sprintf("%s/%s", destination, entry)

		cmd := []string{
			"aws", "s3", "cp",
			entry,
			fileDest,
			"--endpoint-url", endpointURL,
		}

		if m, ok := idx[entry]; ok {
			if m.ContentType != "" {
				cmd = append(cmd, "--content-type", m.ContentType)
			}
			if m.ChecksumSHA256 != "" {
				cmd = append(cmd,
					"--checksum-algorithm", "SHA256",
					"--checksum-sha256", m.ChecksumSHA256,
				)
			}
		}

		awsCli = awsCli.WithExec(cmd)
	}

	_, err = awsCli.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to upload artifacts to %s: %w", destination, err)
	}

	return nil
}

// UploadTree uploads a directory to the bucket under an explicit prefix,
// preserving the directory's internal structure as the key suffix.
//
// Unlike UploadLatest or UploadNightly, which use fixed prefix conventions,
// UploadTree allows the caller to specify any prefix: useful for one off releases,
// OCI registry layouts, or nested directory structures with specifc key paths.
//
// When metadata is supplied, matching files (by relative path) are uploaded
// individually with the specified Content-Type and/or checksum headers.
func (b *Bucketuploader) UploadTree(
	ctx context.Context,

	// Directory to upload — internal structure becomes the key suffix
	artifacts *dagger.Directory,

	// Bucket key prefix. Use "" to upload at the bucket root.
	// +optional
	prefix string,

	// Per-file upload metadata (Content-Type, checksum, etc.).
	// Each entry's Path field should match a relative path inside the artifacts directory.
	// +optional
	metadata []FilePathMetadata,
) error {
	if err := b.upload(ctx, artifacts, prefix, metadata); err != nil {
		return fmt.Errorf("could not upload tree: %w", err)
	}

	return nil
}

// UploadLatest uploads artifacts under both the given version prefix and
// a "latest" prefix, so that the most recent release is always accessible
// at a well-known path.
func (b *Bucketuploader) UploadLatest(
	ctx context.Context,

	// Directory containing build artifacts to upload
	artifacts *dagger.Directory,

	// Version string used as the bucket path prefix (e.g., "v1.2.3")
	version string,

	// Per-file upload metadata (Content-Type, checksum, etc.).
	// Each entry's Path field should match a relative path inside the artifacts directory.
	// +optional
	metadata []FilePathMetadata,
) error {
	if err := b.upload(ctx, artifacts, version, metadata); err != nil {
		return fmt.Errorf("could not upload versioned release artifacts: %w", err)
	}

	if err := b.upload(ctx, artifacts, latest, metadata); err != nil {
		return fmt.Errorf("could not upload latest release artifacts: %w", err)
	}

	return nil
}

// UploadNightly uploads artifacts under the "nightly" prefix.
func (b *Bucketuploader) UploadNightly(
	ctx context.Context,

	// Directory containing build artifacts to upload
	artifacts *dagger.Directory,

	// Per-file upload metadata (Content-Type, checksum, etc.).
	// Each entry's Path field should match a relative path inside the artifacts directory.
	// +optional
	metadata []FilePathMetadata,
) error {
	if err := b.upload(ctx, artifacts, nightly, metadata); err != nil {
		return fmt.Errorf("could not upload nightly artifacts: %w", err)
	}

	return nil
}

// UploadFile uploads a single file to the bucket under an optional path
// prefix. This is useful for standalone files like install scripts.
//
// When metadata is provided, the file is uploaded with the specified
// Content-Type and/or checksum headers.
func (b *Bucketuploader) UploadFile(
	ctx context.Context,

	// The file to upload
	file *dagger.File,

	// Bucket path prefix (e.g., "scripts"). When empty the file is
	// placed at the bucket root.
	// +optional
	prefix string,

	// Upload metadata for this file (Content-Type, checksum, etc.).
	// +optional
	metadata *FileMetadata,
) error {
	dir := dag.Directory().WithFile(".", file)

	// Convert the single FileMetadata into a FilePathMetadata slice
	// keyed by the file's own name so the upload path can look it up.
	var m []FilePathMetadata
	if metadata != nil {
		name, err := file.Name(ctx)
		if err != nil {
			return fmt.Errorf("could not get file name: %w", err)
		}
		m = []FilePathMetadata{{
			Path: name,
			Meta: *metadata,
		}}
	}

	if err := b.upload(ctx, dir, prefix, m); err != nil {
		return fmt.Errorf("could not upload file: %w", err)
	}

	return nil
}
