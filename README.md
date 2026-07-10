# S3 Multipart Upload Example

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


# How Multipart Upload Works

## Create the multipart upload
- Start multipart upload → Get UploadID from S3 via CreateMultipartUpload
- Generate pre-signed URLs for each chunk using UploadID via PresignUploadPart
- Client uploads each part to S3 directly in parallel or sequentially, using PUT and pre-signed URLs
- - Complete multipart upload by calling CompleteMultipartUpload


# API
 ```
Post /api/initialize → with file metadata... Calls CreateMultipartUpload(aws sdk) → returns uploadID.

Post /api/presign_urls → with uploadID + partNumber + filename... Calls PresignUploadPart(aws sdk), generates presigned URLs for each part.
These pre-signed URLs allow the client to upload each chunk directly to S3, securely, without exposing AWS credentials.

Post /api/complete → with list of parts + ETags… Calls CompleteMultipartUpload(aws sdk)
 ```
