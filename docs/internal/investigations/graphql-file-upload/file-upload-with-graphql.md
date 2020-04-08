# File upload with graphql

## Overview

This document discusses how to upload files using graphql and suggests how to handle large data objects.

Vanilla graphql does not support uploading files. However, you can use one of the following workarounds:
- Multipart Request Spec implemented in gqlgen library
- Base64 data encoding
- Middleware which handles uploading files

## Upload files with gqlgen 0.9.0 library

1. Use the following `schema.graphql`:
```graphql
# The `Upload` scalar type represents a multipart file upload.
# It is already implemented in gqlgen library, so we can use it straight away.
scalar Upload

input DocumentationInput {
  data: Upload!
}

type Mutation {
  storeDocumentation(in: DocumentationInput!): String
}

type Query {
  anything: String
}

```

2. Implement resolvers:
```go
type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) StoreDocumentation(ctx context.Context, in DocumentationInput) (*string,error) {
	doc := Documentation{
		ID:string(rand.Int()),
		Data:in.Data,
	}

	b := bytes.NewBuffer(nil)
	io.Copy(b, doc.Data.File)
	out := b.String()
	return &out, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Anything(ctx context.Context) (*string,error){
	return nil,nil
}
```

3. Configure the resolvers:

To make your server recognize that resolver, attach it inside the `main` function. You can also set some additional parameters, such as `UploadMaxMemory` or `UploadMaxSize`.

```go
	exec := fileupload.NewExecutableSchema(fileupload.Config{Resolvers: &fileupload.Resolver{}})

	var mb int64 = 1 << 20
	uploadMaxMemory := handler.UploadMaxMemory(32 * mb)
	uploadMaxSize := handler.UploadMaxSize(50 * mb)

	http.Handle("/query", handler.GraphQL(exec, uploadMaxMemory, uploadMaxSize))
```

4. Test your request:

The curl request accepts a `FILEPATH` variable with the path to the file which you want to send with the request.
```bash
curl localhost:8080/query \
  -F operations='{ "query": "mutation ($file: Upload!) { storeDocumentation(in: {data: $file}) }", "variables": { "file": null } }' \
  -F map='{ "0": ["variables.file"] }' \
  -F 0=@${FILEPATH}
```
For more examples of how to upload files using gqlgen library, go to the reference links section.

## Improvement proposal

We should support multiple ways of uploading large amounts of data, such as documentation.
The `Clob` type, which is to be defined, could look as follows:

```graphql
type Clob {
  type: UploadType!  
  content: String
  file: Upload
}
```

```graphql
enum UploadType {
  FILE
  STRING
}
```

That approach gives the end-user more flexibility on how to upload the data.

## Reference links

- [An article about the possible ways of uploading files](https://medium.freecodecamp.org/how-to-manage-file-uploads-in-graphql-mutations-using-apollo-graphene-b48ed6a6498c)
- [GraphQL multipart request specification](https://github.com/jaydenseric/graphql-multipart-request-spec)
- [gqlgen 0.9.0 library](https://github.com/99designs/gqlgen/tree/v0.9.0)
- [gqlgen file upload example](https://github.com/99designs/gqlgen/tree/v0.9.0/example/fileupload)
