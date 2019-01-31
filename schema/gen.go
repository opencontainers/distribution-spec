package schema

//go:generate esc -private -o=fs.go -modtime=1546544639 -pkg=schema -include=.*\.json$ -ignore=test-fixtures/* .
