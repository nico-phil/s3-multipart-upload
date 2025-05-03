const submitButton = document.getElementById("btn-submit");
const uploadbtn = document.getElementById("btn-s3");

submitButton.addEventListener("click", handleSubmit);

async function handleSubmit(e) {
  e.preventDefault();
  const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk
  const fileInput = document.getElementById("file-input");
  const file = fileInput.files[0];

  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

  // get uploadID from s3
  const uploadID = await getUploadID(file);
  const urls = await generatePresingnUrls(file.name, uploadID, totalChunks);
  const etags = [];

  for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
    const start = chunkIndex * CHUNK_SIZE;
    const end = Math.min(start + CHUNK_SIZE, file.size);
    const chunk = file.slice(start, end);

    const url = urls[chunkIndex];
    const reposeChunk = await uploadChunTos3(url, chunk);
    const etag = reposeChunk.headers.get("ETag");
    etags.push({ etag: etag.replaceAll('"', ""), part_number: chunkIndex + 1 });
  }

  const responsUpload = await completeUpload(file.name, uploadID, etags);
  console.log(responsUpload);
}

async function getUploadID(file) {
  const body = {
    file_name: file.name,
    file_size: file.size,
    file_type: file.type,
  };
  const response = await fetch("http://localhost:3000/initialize", {
    method: "POST",
    body: JSON.stringify(body),
  });

  const jsonResponse = await response.json();
  return jsonResponse.upload_id;
}

async function generatePresingnUrls(fileName, uploadID, totalParts) {
  const body = {
    file_name: fileName,
    upload_id: uploadID,
    total_parts: totalParts,
  };
  const response = await fetch("http://localhost:3000/presigned_urls", {
    method: "POST",
    body: JSON.stringify(body),
  });

  const jsonResponse = await response.json();
  return jsonResponse.urls;
}

async function uploadChunTos3(s3url, chunk) {
  const response = await fetch(s3url, {
    method: "PUT",
    body: chunk,
  });
  return response;
}

async function completeUpload(fileName, uploadID, etags) {
  const body = {
    file_name: fileName,
    upload_id: uploadID,
    parts: etags,
  };
  const response = await fetch("http://localhost:3000/complete", {
    method: "POST",
    body: JSON.stringify(body),
  });

  return response;
}
