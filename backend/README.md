curl http://localhost:8080/stats

curl -X POST http://localhost:8080/upload \
  -F "files=@/full/path/to/your/file.parquet"


  curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "integration-lab:otel-demo", "field": "namespace"}'

  curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "apica-otel-demo-kafka-76f8bdb949-rp65b", "field": "ProcId"}'
