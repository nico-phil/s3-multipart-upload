package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	uploadDir    = "./uploads"
	completeDir  = "./completed"
	maxChunkSize = 5 * 1024 * 1024 // 1MB
)

func main(){
	http.HandleFunc("POST /upload", uploadHandler)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request){
	enableCors(&w)
	err := r.ParseMultipartForm(maxChunkSize + 2 *1024)
	if err != nil {
		http.Error(w, "max chunk exceded", http.StatusBadRequest)
		return
	}

	fileName := r.FormValue("fileName")
	chunkIndex := r.FormValue("chunkIndex")
	totalChunks := r.FormValue("totalChunks")

	index, _ := strconv.Atoi(chunkIndex)
	total, _ := strconv.Atoi(totalChunks)

	// get chunk
	file, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "failed to get chunk", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Printf("Received chunk %d of %d for %s\n", index+1, total, fileName)

	//save chunk to tempory directory
	chunkPath := filepath.Join(uploadDir, fmt.Sprintf("%s.part%d", fileName, index))
	out, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, "Failed to save chunk", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	// fmt.Printf("Received chunk %d of %d for %s\n", index+1, total, fileName)

	// Check if all chunks are received
	complete := true
	for i := 0; i < total; i++ {
		path := filepath.Join(uploadDir, fmt.Sprintf("%s.part%d", fileName, i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			complete = false
			break
		}
	}

	if complete {
		mergeChunks(fileName, total)
		fmt.Printf("✅ Merged all chunks for file: %s\n", fileName)
	}

	w.WriteHeader(http.StatusOK)
	
}

func mergeChunks(fileName string, total int) {
	destPath := filepath.Join(completeDir, fileName)
	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Println("Error creating final file:", err)
		return
	}
	defer destFile.Close()

	for i := 0; i < total; i++ {
		partPath := filepath.Join(uploadDir, fmt.Sprintf("%s.part%d", fileName, i))
		partFile, err := os.Open(partPath)
		if err != nil {
			fmt.Println("Error opening chunk:", err)
			return
		}
		io.Copy(destFile, partFile)
		partFile.Close()
		// os.Remove(partPath) // Clean up
	}
}

func enableCors(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	
}