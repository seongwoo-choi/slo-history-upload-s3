### slo-history-upload-s3

#### 환경 변수

- BUCKET
  - S3 버킷 이름
    - ex) my_bucket
- KEY
  - S3 버킷에 저장할 디렉토리 경로입니다.
    - ex) slo_metric/example_slo
- DAY
  - 현재 날짜에서 7일 전의 SLO History 내역을 가져옵니다.
    - ex) -7, -30, -90
- DD_SITE
  - datadog 사이트를 지정합니다.
    - ex) datadoghq.com
- SLO_ID
  - SLO History 를 가져오려는 SLO_ID 입니다.
    - ex) slo_id_123456789
- DD_API_KEY
  - datadog API 키입니다.
    - ex) 1234567890abcdefg
- DD_APP_KEY
  - datadog APP 키입니다.
    - ex) 1234567890abcdefg
- RECIPIENT
  - SLO 대시보드 링크를 전송할 이메일 주소입니다.
    - ex) "test1@example.com, test2@example.com, test3@example.com"
- DASHBOARD_ID
  - SLO 대시보드 링크를 전송할 대시보드 ID 입니다.
    - ex) 123456


#### 동작

데이터독 SLO_ID 에 해당하는 SLO History 를 가져와서 S3 버킷에 저장합니다. 그 후, AWS SES 를 사용하여 특정 이메일에 데이터독 SLO 대시보드 링크를 전송합니다.