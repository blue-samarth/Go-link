single module : `go mod init github.com/yourusername/myapp`
monorepo with multiple modules : 
```
cd users && go mod init github.com/yourname/project/users && cd ..
cd config && go mod init github.com/yourname/project/config && cd ..
cd services && go mod init github.com/yourname/project/services && cd ..
```
mkdir -p users config services
cd users && go mod init github.com/blue-samarth/go-link/users && cd ..
cd config && go mod init github.com/blue-samarth/go-link/config && cd ..
cd services && go mod init github.com/blue-samarth/go-link/services && cd ..

installations :
`go get github.com/joho/godotenv`