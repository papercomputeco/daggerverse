# github.com/papercomputeco/daggerverse/bucketupload

S3-compat bucket uploading for artifacts via the AWS CLI.


| Function | Description |
|----------|-------------|
| `upload-latest` | Uploads a directory under both a versioned prefix and `latest`. |
| `upload-nightly` | Uploads a directory under the `nightly` prefix. |
| `upload-file` | Uploads a single file, optionally under a path prefix. |


## Constructor arguments

All functions share bucket credentials that are provided once when
constructing the module:

| Argument | Type | Description |
|----------|------|-------------|
| `--endpoint` | `Secret` | Bucket endpoint URL |
| `--bucket` | `Secret` | Bucket name |
| `--access-key-id` | `Secret` | Bucket access key ID |
| `--secret-access-key` | `Secret` | Bucket secret access key |


## Usage

### Upload a versioned release and mirror to latest

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/bucketupload \
  --endpoint env:BUCKET_ENDPOINT \
  --bucket env:BUCKET_NAME \
  --access-key-id env:AWS_ACCESS_KEY_ID \
  --secret-access-key env:AWS_SECRET_ACCESS_KEY \
  upload-latest \
    --artifacts ./dist \
    --version "v1.2.3"
```

### Upload nightly artifacts

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/bucket-upload \
  --endpoint env:BUCKET_ENDPOINT \
  --bucket env:BUCKET_NAME \
  --access-key-id env:AWS_ACCESS_KEY_ID \
  --secret-access-key env:AWS_SECRET_ACCESS_KEY \
  upload-nightly \
    --artifacts ./nightly-build
```

### Upload a single file

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/bucket-upload \
  --endpoint env:BUCKET_ENDPOINT \
  --bucket env:BUCKET_NAME \
  --access-key-id env:AWS_ACCESS_KEY_ID \
  --secret-access-key env:AWS_SECRET_ACCESS_KEY \
  upload-file \
    --file ./install.sh
```
