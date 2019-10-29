# 编译命令
go build -o NAMEOFEXECUTABLE *.go
因为默认go build对象只是所在文件夹下（或者叫包内）单个文件，不会去引用包内其他文件中的方法，需要以这种方式将包内所有文件链接