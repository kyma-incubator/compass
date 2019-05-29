package director

//go:generate protoc --go_out=plugins=grpc:. director.proto
//go:generate protoc --go_out=plugins=grpc:../../gateway/protobuf/ director.proto

