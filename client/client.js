const submitButton = document.getElementById("btn-submit");

submitButton.addEventListener("click", handleSubmit);

async function handleSubmit(e) {
  e.preventDefault();
  const CHUNK_SIZE = 5 * 1024 * 1024; // 1MB per chunk
  const fileInput = document.getElementById("file-input");
  const file = fileInput.files[0];

  if (!file) {
    return;
  }

  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

  for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
    const start = chunkIndex * CHUNK_SIZE;
    const end = Math.min(start + CHUNK_SIZE, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append("chunk", chunk);
    formData.append("fileName", file.name);
    formData.append("chunkIndex", chunkIndex);
    formData.append("totalChunks", totalChunks);

    try {
      const response = await fetch("http://localhost:3000/upload", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`Chunk ${chunkIndex} failed`);
      }

      console.log(`Chunk ${chunkIndex + 1}/${totalChunks} uploaded`);
    } catch (e) {}
  }
}
