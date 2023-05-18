package main

// s3 에 그냥 데이터 업로드 => 아테나 쿼리에서 테이블을 만드는데 s3 에서 데이터를 가져와서 컬럼을 만들 수 있다.
// 퀵사이트에서 s3 를 바로 볼 수 도 있음.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func CreateJSONFile(responseContent []byte) error {
	f, fileCreateErr := os.Create("slo_history.json")
	if fileCreateErr != nil {
		fmt.Println(fileCreateErr)
		return fileCreateErr
	}

	defer f.Close()

	_, fileWriteErr := f.Write(responseContent)
	if fileWriteErr != nil {
		fmt.Println(fileWriteErr)
		return fileWriteErr
	}

	return nil
}

func UploadS3() error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}

	svc := s3.New(sess)

	file, err := os.Open("slo_history.json")
	if err != nil {
		return err
	}
	defer file.Close()

	today := time.Now().Format("2006-01-02")
	bucket := os.Getenv("BUCKET")
	key := os.Getenv("KEY")

	key = fmt.Sprintf("%s/%s/slo_history.json", key, today)

	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
		//ContentType: aws.String(fileType),
	}

	_, err = svc.PutObject(putObjectInput)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("파일이 업로드되었습니다.")

	return nil
}

func GetSloHistory(sloId string, day int) error {
	value := make(map[string]interface{})
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewServiceLevelObjectivesApi(apiClient)
	resp, r, err := api.GetSLOHistory(ctx, sloId, time.Now().AddDate(0, 0, day).Unix(), time.Now().Unix(), *datadogV1.NewGetSLOHistoryOptionalParameters())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceLevelObjectivesApi.GetSLOHistory`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	//fmt.Fprintf(os.Stdout, "Response from `ServiceLevelObjectivesApi.GetSLOHistory`:\n%s\n", responseContent)

	err = json.Unmarshal(responseContent, &value)
	if err != nil {
		return err
	}

	err = CreateJSONFile(responseContent)
	if err != nil {
		return err
	}

	//data := value["data"].(map[string]interface{})
	//monitors := data["monitors"].([]interface{})
	//overall := data["overall"].(map[string]interface{})
	//sliValue := overall["sli_value"].(float64)

	err = UploadS3()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// SLO ID 환경 변수 처리
	sloId := os.Getenv("SLO_ID")
	day, _ := strconv.Atoi(os.Getenv("DAY"))
	err := GetSloHistory(sloId, day)
	if err != nil {
		fmt.Println(err)
	}
}
