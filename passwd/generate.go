package passwd

//go:generate structgen -pkg $GOPACKAGE -out struct.go  -tags plist -repo https://github.com/apple/device-management -path other/passwordhash.yaml -commit release
