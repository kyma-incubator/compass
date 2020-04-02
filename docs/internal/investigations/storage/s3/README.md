# Storage

This document describes using static object storage as a database for the Application Registry.

## Storing Application labels

The idea is to use Apache Parquet for storing pieces of information regarding labels.
We want to have easy access to information and be able to:
    - list all labels for an application
    - query an application by labels
    - list all labels

### S3-based Solution

To achieve querying by groups, the following bucket structure is suggested:

![Bucket structure](./assets/bucket-structure.jpg)

Data redundancy is needed to reduce requests while doing read operations.

In the `data.json` files we could store API, event specs and docs. However, we would need to assure that the latest file is always modified (file versioning). We can also store the files separately, but then the number of requests increases. That's why the cases below assume that we store the data in a single `data.json` file.

**Runtime creation**

- Create a `data.json` file in the `{tenant_id}/runtimes/{runtime_id}` directory (1 request)

For every label of the runtime (N is the number of the runtimes):
- Create the `{key}={value}` file in the `{tenant_id}/runtimes/{runtime_id}/labels` directory (N requests)
- Create the `data.json` file in the `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}` directory (N requests)
- Create the `key=value` file in the `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}/labels` directory (N * N requests)

Total calls: N^2 + 2*N + 1

**Application creation**

Similar to the Runtime creation - replace `runtimes` with `applications`

Total calls: N^2 + 2*N + 1

**Adding a label to an existing Application**

- Create the file `{key}={value}` in the `{tenant_id}/applications/{application_id}` directory (1 request)
- Copy files from the `{tenant_id}/applications/{application_id}` to the `{tenant_id}/labels/{key}/{value}/applications/{application_id}` directory (every file is a separate request; 1 request for data.json + N for labels)

Total calls: N + 2

**Adding a label to an existing Runtime**

- Create the `{key}={value}` file in the `{tenant_id}/runtimes/{runtime_id}` directory (1 request)
- Copy files from the `{tenant_id}/runtimes/{runtime_id}` to the `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}` directory (every file is a separate request; 1 request for data.json + N for labels)

Total calls: N + 2

**List all runtimes**

- List all directories under `{tenant_id}/runtimes` (1 request)

For every runtime (N is the number of the runtimes):
- Get  the`data.json` file from the `{tenant_id}/runtimes/{runtime_id}` directory (N requests)

Total calls: N + 1

**List all applications**

Similar to listing all runtimes - replace `runtimes` with `applications`

Total calls: N + 1

**List application by label**

- List all directories under `{tenant_id}/labels/{key}/{value}/applications` (1 request)

For every application (N is the number of the application):
- Get the `data.json` file from the `{tenant_id}/labels/{key}/{value}/applications/{application_id}` directory (N requests)

Total calls: N + 1

**List all labels**

- List all directories in `{tenant_id}/labels` (1 request)
- Iterate over the list and construct all possible labels ({key}={value})

Total calls: 1

**List all possible values for the label key**

- List all directories under `{tenant_id}/labels/{key}` (1 request)

Total calls: 1

### Pros
- low cost
- region replication
- backup

### Cons
- multiple calls to S3 every time we want to read, write or delete some data

    For example, to query all 10 runtimes for a tenant requires 11 calls to S3:
    -   1 for listing all files in directory
    -   10 requests to download files with data for every runtime.
    -   to add new runtime with 2 labels requires uploading 9 files because of the data redundancy (9 calls - 6 files for labels and 3 files with runtime data)

- complex implementation for adding, removing and querying data
- problems with caching while cross-region replication - while using the cache, even very short-living one, there will be a delay in reading modifications in the `data.json` file if it was accessed before and modified by another client
- No support for transactions (@tgorgol)
    
    For example, if we want to use data redundancy then we need to consider cases where one file write will succeed while another one will fail. This will further increase code complexity.
- No update functionality. (@tgorgol)

    If we decide to store metadata in the `data.json` file, we will have to replace the whole file to be able to update a single value from that file. In our case this wouldn't be a big problem as the `data.json` file is expected to be small, nonetheless, it will increase the complexity of the code, as we will have to first download the existing file, patch it locally and then put the new file in the storage.
    
    One way of solving this issue is to use separate files for each metadata property with its value in the name, so we would have the following directory structure:
    
    ```
    {tenant_id}/
    ├── applications/
    │   ├── {application_id}/
    │   │  ├── name={application_name}
    │   │  ├── description={application_description}
    │   │  ├── annotations/
    │   │  │  ├── {annotation_key} // value in file?
    │   │  │  └── ...
    │   │  ├── apis/
    │   │  │  └── ...
    │   │  ├── labels/
    │   │  │  └── {label_key}={label_value}
    │   │  └── ...
    ...
    ```

   Using the structure, updating a single property would require only replacing a single file. Also, reading application metadata would be faster as most of it would be available after listing the application directory. The downsides to storing this data in a filename are possible length/encoding limitations.

## Cross-region replication

The following section describes a comparison between cloud-storage offerings regarding cross-region replication. All providers listed below have a similar offering, except for the GCP which offers multi-region within only the same continent.

We would like to create multiple replicas in different regions. For example, in the same time we would like to have three replicas in the following regions: Asia, Europe and USA.

### Google Cloud Platform

We can create multi-regional GCS buckets, but this functionality is limited to the following options:

- United States (multiple regions in the United States) 
- European Union (multiple regions in the European Union) 
- Asia (multiple regions in Asia)
- eur4 (Finland and Netherlands) 
- nam4 (Iowa and South Carolina) 

We cannot replicate a single bucket across multiple regions like Asia, USA and Europe. Therefore, it doesn't meet our requirements. 

### AWS S3

Amazon has [Cross-region replication for S3 buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html). One bucket would be used to write, and others to read. It is limited to two regions at the same time (1:1 bucket mapping). It doesn't meet our requirements. 

### Azure Blob Storage

Microsoft Azure has [Read-Access Geo-Redundant Storage (RA-GRS)](https://docs.microsoft.com/pl-pl/azure/storage/common/storage-redundancy) available, but it is limited to only two regions at the same time. As a result, it doesn't meet our requirements.

## Summary

There are many [discussions](https://www.quora.com/How-can-we-use-Amazon-S3-as-a-database) whether using S3 as a database is a good idea. It might work in some specific use cases, but generally, it is not suggested to do so.

In our case using S3 will make the Registry implementation much more difficult. Doing so, many requests per single read/write operation will mean that every request to the Registry will be slow - much slower than a typical (No)SQL database.

Also, the cross-region replication is very limited in the top cloud providers' offerings. There is no cloud provider from the three investigated that meets our requirements.

We should investigate database cloud solutions. In my opinion, in our use case, the NoSQL database will fit more than SQL ones. We could investigate solutions such as [Cloud Firestone](https://cloud.google.com/firestore/) or [Amazon DynamoDB](https://aws.amazon.com/dynamodb/).
