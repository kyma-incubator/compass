# File upload with graphql

## Overview

This document discuss the issue of uploading files with graphql. \
As vanilla graphql does not support uploading files, there are several workarounds

### File uploading with gqlgen 0.9.0 library
<details>
<summary>Example of schema.graphql</summary>
<p>

```graphql
#"The `Upload` scalar type represents a multipart file upload."
scalar Upload

#"The `File` type, represents the response of uploading a file."
type File {
    id: Int!
    name: String!
    content: String!
}

#"The `UploadFile` type, represents the request for uploading a file with certain payload."
input UploadFile {
    id: Int!
    file: Upload!
}

#"The `Query` type, represents all of the entry points into our object graph."
type Query {
    empty: String!
}

#"The `Mutation` type, represents all updates we can make to our data."
type Mutation {
    singleUpload(file: Upload!): File!
    singleUploadWithPayload(req: UploadFile!): File!
    multipleUpload(files: [Upload!]!): [File!]!
    multipleUploadWithPayload(req: [UploadFile!]!): [File!]!
}
```
</p>
</details>
<details>
<summary>Resolvers</summary>
<p>

```graphql

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) SingleUpload(ctx context.Context, file graphql.Upload) (*File, error) {
	return r.SingleUpload(ctx, file)
}
func (r *mutationResolver) SingleUploadWithPayload(ctx context.Context, req UploadFile) (*File, error) {
	return r.SingleUploadWithPayload(ctx, req)
}
func (r *mutationResolver) MultipleUpload(ctx context.Context, files []*graphql.Upload) ([]*File, error) {
	return r.MultipleUpload(ctx, files)
}
func (r *mutationResolver) MultipleUploadWithPayload(ctx context.Context, req []*UploadFile) ([]*File, error) {
	return r.MultipleUploadWithPayload(ctx, req)
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Empty(ctx context.Context) (string, error) {
	return r.Empty(ctx)
}
```
</p>
</details>
<details>
<summary>Implementation of resolvers</summary>
<p>

```go
func getResolver() *go_graphql_demo.Stub {
	resolver := &go_graphql_demo.Stub{}

	resolver.MutationResolver.SingleUpload = func(ctx context.Context, file graphql.Upload) (*go_graphql_demo.File, error) {
		content, err := ioutil.ReadAll(file.File)
		if err != nil {
			return nil, err
		}
		return &go_graphql_demo.File{
			ID:      1,
			Name:    file.Filename,
			Content: string(content),
		}, nil
	}
	resolver.MutationResolver.SingleUploadWithPayload = func(ctx context.Context, req go_graphql_demo.UploadFile) (*go_graphql_demo.File, error) {
		content, err := ioutil.ReadAll(req.File.File)
		if err != nil {
			return nil, err
		}
		return &go_graphql_demo.File{
			ID:      1,
			Name:    req.File.Filename,
			Content: string(content),
		}, nil
	}
	resolver.MutationResolver.MultipleUpload = func(ctx context.Context, files []*graphql.Upload) ([]*go_graphql_demo.File, error) {
		if len(files) == 0 {
			return nil, errors.New("empty list")
		}
		var resp []*go_graphql_demo.File
		for i := range files {
			content, err := ioutil.ReadAll(files[i].File)
			if err != nil {
				return []*go_graphql_demo.File{}, err
			}
			resp = append(resp, &go_graphql_demo.File{
				ID:      i + 1,
				Name:    files[i].Filename,
				Content: string(content),
			})
		}
		return resp, nil
	}
	resolver.MutationResolver.MultipleUploadWithPayload = func(ctx context.Context, req []*go_graphql_demo.UploadFile) ([]*go_graphql_demo.File, error) {
		if len(req) == 0 {
			return nil, errors.New("empty list")
		}
		var resp []*go_graphql_demo.File
		for i := range req {
			content, err := ioutil.ReadAll(req[i].File.File)
			if err != nil {
				return []*go_graphql_demo.File{}, err
			}
			resp = append(resp, &go_graphql_demo.File{
				ID:      i + 1,
				Name:    req[i].File.Filename,
				Content: string(content),
			})
		}
		return resp, nil
	}
	return resolver
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
resolver := getResolver()
exec := fileupload.NewExecutableSchema(fileupload.Config{Resolvers: resolver})

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
curl localhost:8087/query \
  -F operations='{ "query": "mutation ($file: Upload!) { singleUpload(file: $file) { id, name, content } }", "variables": { "file": null } }' \
  -F map='{ "0": ["variables.file"] }' \
  -F 0=@${FILEPATH}

```

</p>
</details>

For more examples check reference links section.

#### Reference links

- [Article about file upload possibilities](https://medium.freecodecamp.org/how-to-manage-file-uploads-in-graphql-mutations-using-apollo-graphene-b48ed6a6498c)
- [gqlgen 0.9.0 library](https://github.com/99designs/gqlgen/tree/v0.9.0)
- [gqlgen file upload example](https://github.com/99designs/gqlgen/tree/v0.9.0/example/fileupload)