# Apache Parquet

The idea is to use Apache Parquet for storing group relations between Runtime and the Application.
Apache Parquet is a column-oriented data storage format, which could be used to hold the needed pieces of information. It would be stored on the root of the S3 bucket. 

The whole structure on S3 would be as follows:

```
{tenant_id}/
├── applications/
│   ├── {application_id}/
│   │  └── data.json
├── runtimes/
│   ├── {application_id}/
│   │  └── data.json
└── db.parquet
```

To minimize the file size, the `db.parquet` file would store all details regarding labels for runtimes and applications. We would do queries for:
- all labels
- all applications with a specific label
- all runtimes with a specific label
- all labels with a specific key

However, all read/write operations would need multiple S3 requests to handle them. For example:

To read all runtimes (when we have 10 of them), we would need to:
- fetch list of the `runtimes` directory (1 request)
- for every runtime, we would need to download the `data.json` file (10 requests)
- download the `db.parquet` file (1 request), query labels for all runtimes and merge the above results with data

We could use a similar approach to the one described [here](./README.md), by writing labels for runtimes and applications as files under the `{tenant_id}/{applications or runtimes}/{application_id or runtime_id}/` directory.

### Pros

1. **Efficient data store**

    Very good compression. For example, if the CSV file has 1 TB, the data stored in Parquet file has 130 GB.

1. **Good performance for reading data**
    
    Fetching specific column values doesn't need reading the entire raw data.

### Cons

1. **Doing queries**

    For Golang, we have [parquet-go](https://github.com/xitongsys/parquet-go) library, which handles the reading and writing of the `parquet` file.

    In order to do queries in Apache Parquet, we would need a query engine, such as [Apache Drill](https://drill.apache.org/), which is written in Java.

    If we wanted to avoid using this engine. We would need to rethink how to structure the data, to be able to cover all our use cases with just the ability to read a specific column.

1. **Concurrent write into the same file**

    We could enable bucket files versioning (for example, in [GCS](https://cloud.google.com/storage/docs/gsutil/addlhelp/ObjectVersioningandConcurrencyControl)). However, we would need to implement a mechanism to always fetch the latest version of the file and retry saving until it is successful.
 
1. **Time-consuming operations**

    For example, every time a user adds an application to a group, we would need to download the `parquet` file, modify it and then upload the modified version. We can't implement caching as we always need the latest version of this file, especially when we will use Cross-region replicas.

    However, there is S3 Select available on AWS for [Parquet](https://aws.amazon.com/about-aws/whats-new/2018/09/amazon-s3-announces-new-features-for-s3-select/), which would eliminate the need to download the whole file for read-only operations. Minio supports S3 Select. Still, writing will always mean that we need to download the file.

### Summary

The Apache Parquet is a good solution for a different use case - when storing a big amount of data with complex structure. In our case, we can store labels using standard S3 capabilities: creating and removing files and directories. We can create a directory structure on S3 to enable querying by groups.
