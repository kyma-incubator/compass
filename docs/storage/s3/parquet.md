# Apache Parquet

There is idea to use Apache Parquet for storing group relations between Runtime and the Application.
Apache Parquet is a column-oriented data storage format, which could be used to hold the needed pieces of information. It would be stored on the root of the S3 bucket. 

### Pros

1. **Efficient data store**

    Very good compression. For example, if CSV file is as big as 1 TB, the data stored in Parquet file has 130 GB.

1. **Good performance**
    
    Fetching specific column values doesn't need reading the entire raw data.

### Cons

1. **Doing queries**

    For Golang, we have [parquet-go](https://github.com/xitongsys/parquet-go) library, which handles read and writes of the `parquet` file.

    In order to do queries in Apache Parquet, we would need a query engine, such as [Apache Drill](https://drill.apache.org/), which is written in Java.

    If we wanted to avoid this engine, in our use case, we would need to define group names and application names as columns.

1. **Concurrent write into the same file**

    We could enable bucket files versioning (for example, in [GCS](https://cloud.google.com/storage/docs/gsutil/addlhelp/ObjectVersioningandConcurrencyControl)). However, we would need to implement a mechanism to always fetch the latest version of the file and retry saving until success.
 
1. **Time-consuming operations related to labels**

    For example, every time the user adds an application to a group, we would need to download `parquet` file, modify it and then upload modified version. We can't implement caching as we always need latest version of this file, especially when we will use Cross-region replicas.

    However, there is S3 Select available on AWS for [Parquet](https://aws.amazon.com/about-aws/whats-new/2018/09/amazon-s3-announces-new-features-for-s3-select/), which would eliminate the need to download the whole file for read-only operations.

### Summary

The Apache Parquet is good in different use case, when storing big amount of data with complex structure. In our case, we can store labels using standard S3 capabilities: creating and removing files and directories. We can create a directory structure on S3 to enable querying by groups.