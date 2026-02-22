package main

import (
	"context"
	"fmt"
	"path"

	"dagger/bucketuploader/internal/dagger"
)

const (
	nightly = "nightly"
	latest  = "latest"
)

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

// upload syncs a directory to the bucket under the given prefix.
func (b *Bucketuploader) upload(
	ctx context.Context,
	artifacts *dagger.Directory,
	prefix string,
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

// UploadTree uploads a directory to the bucket under an explicit prefix,
// preserving the directory's internal structure as the key suffix.
//
// Unlike UploadLatest or UploadNightly, which use fixed prefix conventions,
// UploadTree allows the caller to specify any prefix: useful for one off releases,
// OCI registry layouts, or nested directory structures with specifc key paths.
func (b *Bucketuploader) UploadTree(
	ctx context.Context,

	// Directory to upload — internal structure becomes the key suffix
	artifacts *dagger.Directory,

	// Bucket key prefix. Use "" to upload at the bucket root.
	// +optional
	prefix string,
) error {
	if err := b.upload(ctx, artifacts, prefix); err != nil {
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
) error {
	if err := b.upload(ctx, artifacts, version); err != nil {
		return fmt.Errorf("could not upload versioned release artifacts: %w", err)
	}

	if err := b.upload(ctx, artifacts, latest); err != nil {
		return fmt.Errorf("could not upload latest release artifacts: %w", err)
	}

	return nil
}

// UploadNightly uploads artifacts under the "nightly" prefix.
func (b *Bucketuploader) UploadNightly(
	ctx context.Context,

	// Directory containing build artifacts to upload
	artifacts *dagger.Directory,
) error {
	if err := b.upload(ctx, artifacts, nightly); err != nil {
		return fmt.Errorf("could not upload nightly artifacts: %w", err)
	}

	return nil
}

// UploadFile uploads a single file to the bucket under an optional path
// prefix. This is useful for standalone files like install scripts.
func (b *Bucketuploader) UploadFile(
	ctx context.Context,

	// The file to upload
	file *dagger.File,

	// Bucket path prefix (e.g., "scripts"). When empty the file is
	// placed at the bucket root.
	// +optional
	prefix string,
) error {
	dir := dag.Directory().WithFile(".", file)
	if err := b.upload(ctx, dir, prefix); err != nil {
		return fmt.Errorf("could not upload file: %w", err)
	}

	return nil
}
