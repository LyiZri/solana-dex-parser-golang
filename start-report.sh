pkill -f go-report-processor

go build -o go-report-processor src/main.go

rm nohup.out

nohup ./go-report-processor > nohup.out 2>&1 &