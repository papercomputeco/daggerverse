package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/bucketuploader/internal/dagger"
)

const (
	nightly = "nightly"
	latest  = "latest"
)

// Bucketuploader provides bucket upload artifact capabilities.
// It expects an S3-compatible bucket and uses rclone for uploads.
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

// rcloneContainer returns a container with rclone configured for the bucket.
// The "r2" remote is set up entirely through environment variables.
func (b *Bucketuploader) rcloneContainer(
	ctx context.Context,
	artifacts *dagger.Directory,
) (*dagger.Container, string, error) {
	bucketName, err := b.Bucket.Plaintext(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get bucket name: %w", err)
	}

	endpointURL, err := b.Endpoint.Plaintext(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get endpoint: %w", err)
	}

	ctr := dag.Container().
		From("rclone/rclone:latest").
		WithSecretVariable("RCLONE_CONFIG_R2_ACCESS_KEY_ID", b.AccessKeyID).
		WithSecretVariable("RCLONE_CONFIG_R2_SECRET_ACCESS_KEY", b.SecretAccessKey).
		WithEnvVariable("RCLONE_CONFIG_R2_TYPE", "s3").
		WithEnvVariable("RCLONE_CONFIG_R2_PROVIDER", "Cloudflare").
		WithEnvVariable("RCLONE_CONFIG_R2_ENDPOINT", endpointURL).
		WithMountedDirectory("/artifacts", artifacts).
		WithWorkdir("/artifacts")

	return ctr, bucketName, nil
}

// upload copies files from the directory into the bucket under the given
// prefix using rclone. When no per-file metadata is provided, a bulk
// "rclone copy" is used. When metadata is present, files with metadata
// entries are uploaded individually via "rclone copyto" with the
// appropriate --header-upload flags; remaining files use the bulk path.
func (b *Bucketuploader) upload(
	ctx context.Context,
	artifacts *dagger.Directory,
	prefix string,
	metadata []FilePathMetadata,
) error {
	ctr, bucketName, err := b.rcloneContainer(ctx, artifacts)
	if err != nil {
		return err
	}

	dest := fmt.Sprintf("r2:%s/%s", bucketName, prefix)

	if len(metadata) == 0 {
		// Fast path: no per-file metadata, bulk copy.
		_, err = ctr.
			WithExec([]string{
				"rclone", "copy", ".",
				dest,
				"--transfers", "4",
				"--s3-chunk-size", "100M",
			}).
			Sync(ctx)
		if err != nil {
			return fmt.Errorf("failed to upload artifacts to %s: %w", dest, err)
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

	// Upload each file individually via "rclone copyto".
	// Glob returns directory entries with a trailing slash — skip them.
	for _, entry := range entries {
		if strings.HasSuffix(entry, "/") {
			continue
		}

		fileDest := fmt.Sprintf("%s/%s", dest, entry)

		cmd := []string{
			"rclone", "copyto",
			entry,
			fileDest,
			"--s3-chunk-size", "100M",
		}

		if m, ok := idx[entry]; ok {
			if m.ContentType != "" {
				cmd = append(cmd,
					"--header-upload",
					fmt.Sprintf("Content-Type: %s", m.ContentType),
				)
			}
			if m.ChecksumSHA256 != "" {
				cmd = append(cmd,
					"--header-upload",
					fmt.Sprintf("x-amz-checksum-sha256: %s", m.ChecksumSHA256),
				)
			}
		}

		ctr = ctr.WithExec(cmd)
	}

	_, err = ctr.Sync(ctx)
	if err != nil {
		return fmt.Errorf("failed to upload artifacts to %s: %w", dest, err)
	}

	return nil
}

// UploadTree uploads a directory to the bucket under an explicit prefix,
// preserving the directory's internal structure as the key suffix.
//
// Unlike UploadLatest or UploadNightly, which use fixed prefix conventions,
// UploadTree allows the caller to specify any prefix: useful for one off releases,
// OCI registry layouts, or nested directory structures with specific key paths.
//
// When metadata is supplied, matching files (by relative path) are uploaded
// with the specified Content-Type and/or pre-computed SHA-256 checksum.
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
			Path:           name,
			ChecksumSHA256: metadata.ChecksumSHA256,
			ContentType:    metadata.ContentType,
		}}
	}

	if err := b.upload(ctx, dir, prefix, m); err != nil {
		return fmt.Errorf("could not upload file: %w", err)
	}

	return nil
}
