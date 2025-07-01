cd $1
go mod init $1
echo "require local/common v0.0.0" >> go.mod 
echo "replace local/common => ../common" >> go.mod 
go mod tidy
go build
