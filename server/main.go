package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	s3Client *s3.Client
	BUCKET_NAME = ""
)

type Presigner struct {
	PresignClient *s3.PresignClient
}

func (presigner Presigner) GetObject(ctx context.Context, bucketName string, objectKey string, lifeTimeSecs int)(*v4.PresignedHTTPRequest, error){
	request, err := presigner.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(objectKey),
	}, func(opts *s3.PresignOptions){
		opts.Expires = time.Duration(lifeTimeSecs * int(time.Second))
	})

	if err != nil {
		log.Printf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}

func(presigner Presigner) PutObject(ctx context.Context, bucketName, objectName string, lifeTimeSecs int64)(*v4.PresignedHTTPRequest, error){
	req, err := presigner.PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(objectName),
	}, func(opts *s3.PresignOptions){
		opts.Expires = time.Duration(lifeTimeSecs * int64(time.Second))
	})

	if err != nil {
		return &v4.PresignedHTTPRequest{}, err
	}

	return req, err
}

type BucketBasics struct {
	S3Client *s3.Client
}

func mutlipartUoload(fileName string)(*s3.CreateMultipartUploadOutput, error){
	output , err := s3Client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket:  aws.String(BUCKET_NAME),
		Key:  aws.String(fileName),
	})

	if err != nil {
		return &s3.CreateMultipartUploadOutput{}, err
	}

	return output, nil
}

// the client this req
func(p Presigner) GetPresingUrlsForEachPart(uploadID string, fileName string, partNumber int32)(*v4.PresignedHTTPRequest, error){

	presigndUrl, err := p.PresignClient.PresignUploadPart(context.Background(), &s3.UploadPartInput {
		Bucket:  aws.String(BUCKET_NAME),
		Key:  aws.String(fileName),
		UploadId: aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(10 * time.Minute)
	})

	if err != nil {
		return &v4.PresignedHTTPRequest{}, err
	}

	fmt.Println("presinged", presigndUrl.URL)

	return  presigndUrl, nil

}






func main(){
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	s3Client = s3.NewFromConfig(cfg)

	
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)


	mux.HandleFunc("POST /initialize", initializeMultiiPartUploader)

	mux.HandleFunc("POST /presigned_urls", generatePresigedUrl)

	mux.HandleFunc("POST /complete", completeMutipleUpload)

	http.ListenAndServe(":3000", mux)
	

}

func completeMutipleUpload(w http.ResponseWriter, r *http.Request){
	enableCors(&w)
	type Part struct {
		ETag string `json:"etag"`
		PartNumber int64  `json:"part_number"`
	}
	var input struct {
		UploadID string `json:"upload_id"`
		FileName string `json:"file_name"`
		Parts []Part `json:"parts"`
	}
	err := readJson(r, &input)
	if err != nil {
		http.Error(w, fmt.Sprintf("mal formed json: %v", err), http.StatusBadRequest)
		return
	}

	var completedParts []types.CompletedPart
	for _, part := range input.Parts {
		completedParts = append(completedParts, types.CompletedPart{
			ETag:       aws.String(part.ETag),
			PartNumber: aws.Int32(int32(part.PartNumber)),
		})
	}

	 completeOutput, err := s3Client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:  aws.String("invoice-s3-bucket-dev"),
		Key:  aws.String(input.FileName),
		UploadId: aws.String(input.UploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})

	

	if err != nil {
		http.Error(w, fmt.Sprintf("error completing upload: %v", err), http.StatusInternalServerError)
	}


	data := map[string]any{
		"input": input,
		"output": completeOutput,
	}
	js, _ := json.MarshalIndent(data, "", "\t")
	// fmt.Fprintf(w, "finised uploading, key: %s", *c.Key)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(js))
}

func initializeMultiiPartUploader(w http.ResponseWriter, r *http.Request){
	enableCors(&w)

	var input struct {
		FileName string `json:"file_name"`
		// FileSize int `json:"file_size"`
		// FileType string `json:"file_type"`
	}

	err := readJson(r, &input)
	if err != nil {
		http.Error(w, "mal formed json", http.StatusBadRequest)
		return
	}

	output, err := mutlipartUoload(input.FileName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	
	data := map[string]string {
		"upload_id": *output.UploadId,
	}

	js, _ := json.MarshalIndent(data, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(js)
}

func generatePresigedUrl(w http.ResponseWriter, r *http.Request){
	enableCors(&w)
	var input struct {
		UploadID string `json:"upload_id"`
		FileName string `json:"file_name"`
		TotalParts int `json:"total_parts"`

	}

	_ = readJson(r, &input)
	
	
	presigner := Presigner{PresignClient: s3.NewPresignClient(s3Client) }
	urls := []string{}
	for i := 1; i <= input.TotalParts; i++ {

		p, err := presigner.GetPresingUrlsForEachPart(input.UploadID, input.FileName, int32(i))
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot generate presinged urls: %v", err), http.StatusInternalServerError)
			return
		}
		urls = append(urls, p.URL)
	}
	
	data := map[string]any {
		"urls": urls,
	}
	js, _ := json.MarshalIndent(data , "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(js)
}



func hello(w http.ResponseWriter, r *http.Request){
	fmt.Println("hello")

}

func enableCors(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	
}

func readJson(r *http.Request, dst any) error{
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&dst)
	if err != nil {
		return err
		
	}

	return nil
}