# File upload with graphql

## Overview

This document discuss the issue of uploading files with graphql. \
As vanilla graphql does not support uploading files, there are several workarounds

### File uploading with gqlgen 0.9.0 library
<details>
<summary>Example of schema.graphql</summary>
<p>

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
</p>
</details>

<details>
<summary>Implementation of resolvers</summary>
<p>

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


</p>
</details>

<details>
<summary> Resolver configuration </summary>
<p>

To make your server recognise that resolver, attach it inside `main` function. \
You can also set some parameters like `UploadMaxMemory` or `UploadMaxSize`.

```go
	exec := fileupload.NewExecutableSchema(fileupload.Config{Resolvers: &fileupload.Resolver{}})

	var mb int64 = 1 << 20
	uploadMaxMemory := handler.UploadMaxMemory(32 * mb)
	uploadMaxSize := handler.UploadMaxSize(50 * mb)

	http.Handle("/query", handler.GraphQL(exec, uploadMaxMemory, uploadMaxSize))
```

</p>
</details>

<details>
<summary> Test request </summary>
<p>

The curl request accepts a `FILEPATH` variable with the path to file which you want to send with the request
```bash
curl localhost:8080/query \
  -F operations='{ "query": "mutation ($file: Upload!) { storeDocumentation(in: {data: $file}) }", "variables": { "file": null } }' \
  -F map='{ "0": ["variables.file"] }' \
  -F 0=@${FILEPATH}
```

</p>
</details>

For more examples check reference links section.

### Improvement proposal

We should support multiple ways of uploading large amounts of data (for example documentation).
The `Clob` type which is in to be defined state, could look like that:

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

#### Reference links

- [Article about file upload possibilities](https://medium.freecodecamp.org/how-to-manage-file-uploads-in-graphql-mutations-using-apollo-graphene-b48ed6a6498c)
- [gqlgen 0.9.0 library](https://github.com/99designs/gqlgen/tree/v0.9.0)
- [gqlgen file upload example](https://github.com/99designs/gqlgen/tree/v0.9.0/example/fileupload)