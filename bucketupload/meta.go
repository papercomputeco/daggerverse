package main

import "strings"

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

// FilePathMetadata pairs a relative file path with upload metadata.
// Used by directory-based upload methods to apply per-file headers.
type FilePathMetadata struct {
	// Relative path of the file inside the artifacts directory
	// (e.g., "bin/my-binary").
	Path string

	// Base64-encoded SHA-256 checksum of the file contents.
	// When set, the x-amz-checksum-sha256 header is sent with the upload.
	// +optional
	ChecksumSHA256 string

	// MIME content type for the file (e.g., "application/octet-stream").
	// When set, the Content-Type header is sent with the upload.
	// +optional
	ContentType string
}

// WithChecksumSHA256 sets the base64-encoded SHA-256 checksum that will be
// sent as the x-amz-checksum-sha256 header during upload.
func (m *FileMetadata) WithChecksumSHA256(
	// Base64-encoded SHA-256 checksum
	checksum string,
) *FileMetadata {
	m.ChecksumSHA256 = checksum
	return m
}

// WithContentType sets the MIME content type that will be sent as the
// Content-Type header during upload.
func (m *FileMetadata) WithContentType(
	// MIME content type (e.g., "application/octet-stream")
	contentType string,
) *FileMetadata {
	m.ContentType = contentType
	return m
}

// WithChecksumSHA256 sets the base64-encoded SHA-256 checksum that will be
// sent as the x-amz-checksum-sha256 header during upload.
func (pm *FilePathMetadata) WithChecksumSHA256(
	// Base64-encoded SHA-256 checksum
	checksum string,
) *FilePathMetadata {
	pm.ChecksumSHA256 = checksum
	return pm
}

// WithContentType sets the MIME content type that will be sent as the
// Content-Type header during upload.
func (pm *FilePathMetadata) WithContentType(
	// MIME content type (e.g., "application/octet-stream")
	contentType string,
) *FilePathMetadata {
	pm.ContentType = contentType
	return pm
}

// NewFileMetadata returns a new empty NewFileMetadata instance.
// Use the With* methods to set individual fields:
//
//	meta := uploader.NewFileMetadata().
//	    WithContentType("application/gzip").
//	    WithChecksumSHA256("base64hash...")
func (b *Bucketuploader) NewFileMetadata() *FileMetadata {
	return &FileMetadata{}
}

// NewFilePathMetadata returns a new NewFilePathMetadata for the given relative path.
// Use the With* methods to set individual fields:
//
//	pm := uploader.NewFilePathMetadata("bin/my-binary").
//	    WithContentType("application/gzip").
//	    WithChecksumSHA256("base64hash...")
func (b *Bucketuploader) NewFilePathMetadata(
	// Relative path of the file inside the artifacts directory
	filePath string,
) *FilePathMetadata {
	return &FilePathMetadata{Path: filePath}
}

// metadataIndex maps cleaned relative file paths to their metadata.
type metadataIndex map[string]FilePathMetadata

// buildMetadataIndex creates a lookup map from a slice of FilePathMetadata,
// normalizing paths by stripping leading "./" and "/" prefixes.
func buildMetadataIndex(metadata []FilePathMetadata) metadataIndex {
	idx := make(metadataIndex, len(metadata))
	for _, m := range metadata {
		key := strings.TrimPrefix(strings.TrimPrefix(m.Path, "./"), "/")
		idx[key] = m
	}
	return idx
}
