для использования как комманды собери бинарник:

go build -o scaffold main.go

Как запускать
go run . create task-client --postgres --kafka

или если бинарь уже установлен:

scaffold create task-client --postgres --kafka
