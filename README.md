# S3 Multipart Upload

This project demonstrates how to upload large files to Amazon S3 using the multipart upload API.

Multipart upload splits a file into smaller parts, uploads those parts independently, and then combines them into a single S3 object. This approach is useful for large files because it improves reliability and allows multiple parts to be uploaded concurrently.


# Features
Initiates an S3 multipart upload
Splits a file into multiple parts
Uploads parts concurrently
Collects uploaded part ETags
Completes the multipart upload
Aborts the upload when an error occurs
Supports configurable part size and concurrency


# Create the multipart upload
- Start multipart upload → Get UploadID from S3 via CreateMultipartUpload
- Generate pre-signed URLs for each chunk using UploadID via PresignUploadPart
- Client uploads each part to S3 directly in parallel or sequentially, using PUT and pre-signed URLs
- Complete multipart upload by calling CompleteMultipartUpload


# API

Start multipart upload → Get UploadID from S3
 ```
Post /api/initialize -> returns uploadID.
{
    file_name
    file_size 
    file_type
}
```

Generate pre-signed URLs for each chunk using UploadID
```
Post /api/presign_urls return []presing_url
{
    upload_id
    file_name
    total_parts
}

```

Complete multipart upload
```
Post /api/complete 
{    
    upload_id
    file_name
    parts []Part
}

Part{
    etag
    part_number
}

```

# Run The server

clone the repo
```
git clone https://github.com/nico-phil/s3-multipart-upload.git
cd server
go run main.go
```
