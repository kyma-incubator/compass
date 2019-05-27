# Storage

Description of using static object storage as database for Application Registry.

## Storing Application labels

There were an idea to use Apache Parquet for storing the pieces of information regarding labels.
We want to have an easy access to the pieces of information:
    - list all labels for application 
    - query application by label
    - list all labels

### S3-based Solution

To achieve querying by groups, the following bucket structure is suggested:

![Bucket structure](./assets/bucket-structure.jpg)

Data redundancy is needed to reduce requests while doing read operations.

In `data.json` files we could store API, event specs and docs. However, we would need to assure that always the latest file is modified (file versioning). We can also store the files separately, but then the number of requests increases. That's why the cases below assumes that we store the data in single `data.json` file.

**Runtime creation**

- Create `data.json` file in `{tenant_id}/runtimes/{runtime_id}` directory (1 request)

For every label of the runtime (N is the number of the runtimes):
- Create `{key}={value}` file for every label in `{tenant_id}/runtimes/{runtime_id}/labels` (N requests)
- Create `data.json` file in `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}` directory (N requests)
- Create `key=value` file in `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}/labels` directory (N * N requests)

Total calls: N^2 + 2*N + 1

**Application creation**

Similar to Runtime creation - replace `runtimes` with `applications`

Total calls: N^2 + 2*N + 1

**Adding label to an existing Application**

- Create file `{key}={value}` in `{tenant_id}/applications/{application_id}` directory (1 request)
- Copy files from `{tenant_id}/applications/{application_id}` to `{tenant_id}/labels/{key}/{value}/applications/{application_id}` directory (every file is a separate request; 1 request for data.json + N for labels)

Total calls: N + 2

**Adding label to an existing Runtime**

- Create file `{key}={value}` in `{tenant_id}/runtimes/{runtime_id}` directory (1 request)
- Copy files from `{tenant_id}/runtimes/{runtime_id}` to `{tenant_id}/labels/{key}/{value}/runtimes/{runtime_id}` directory (every file is a separate request; 1 request for data.json + N for labels)

Total calls: N + 2

**List all runtimes**

- List all directories under `{tenant_id}/runtimes` (1 request)

For every runtime (N is the number of the runtimes):
- Get `data.json` file from `{tenant_id}/runtimes/{runtime_id}` directory (N requests)

Total calls: N + 1

**List all applications**

Similar to listing all runtimes - replace `runtimes` with `applications`

Total calls: N + 1

**List application by label**

- List all directories under `{tenant_id}/labels/{key}/{value}/applications` (1 request)

For every application (N is the number of the application):
- Get `data.json` file from `{tenant_id}/labels/{key}/{value}/applications/{application_id}` directory (N requests)

Total calls: N + 1

**List all labels**

- List all directories in `{tenant_id}/labels` (1 request)
- Iterate over the list and construct all possible labels ({key}={value})

Total calls: 1

**List all possible values for label key**

- List all directories under `{tenant_id}/labels/{key}` (1 request)

Total calls: 1

### Pros
- low cost
- region replication
- backup

### Cons
- multiple calls to S3 every time we want to read, write or delete some data. 

    For example, to query all 10 runtimes for tenant, we would need to do 11 calls to S3:
    -   1 for listing all files in directory
    -   10 requests to download files with data for every runtime.
    -   to Add new runtime with 2 labels, we would need to upload 9 files because of the data redundancy (9 calls - 6 files for labels and 3 files with runtime data)

- complex implementation for adding, removing and querying data
- problems with caching while cross-region replication - while using cache, even very short-living one, there will be a delay for reading modifications in `data.json` file, if it were accessed before and modified by other client
- No support for transactions (@tgorgol)
    
    For example, if we want to use data redundancy then we need to consider cases where one file write will succeed while other one will fail. This will further increase code complexity.
- No update functionality. (@tgorgol)

    If we decide to store metadata in `data.json` file then to update a single value from that file we have to replace whole file. In our case this wouldn't be a big problem as the `data.json` file is expected to be small, nonetheless it will increase complexity of code, as we will have to first download current file, patch it locally and then put the new file in the storage.
    
    One way of solving this issue is to use separate files for each metadata property with it's value in name, so we would have a directory structure like that:
    
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

   that way updating single property would require only replacing a single file, also, reading application metadata would be faster as most of it would be available after listing the application directory. The downsides to storing this data in filename are possible length/encoding limitations.

## Cross-region replication

The following section describes a comparison between cloud-storage offerings, regarding the cross-region replication. All providers listed below have similiar offering, except the GCP, which has multi-region within only the same continent.

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

On Microsoft Azure, there is available [Read-Access Geo-Redundant Storage (RA-GRS)](https://docs.microsoft.com/pl-pl/azure/storage/common/storage-redundancy), but it is limited to only two regions at the same time. In a result, it doesn't meet our requirements.

## Summary

There are many [discussions](https://www.quora.com/How-can-we-use-Amazon-S3-as-a-database) whether using S3 as database is a good idea. It might work in some specific use cases, but generally it is not suggested to do so.

In our case using S3 will make the Registry implementation much more difficult. Doing so many request per single read/write operation will mean that every request to the Registry will be slow - much slower than typical (No)SQL database.

Also, the cross-region replication is very limited in the top cloud providers. There is no cloud provider from three investigated that meets our requirements.

We should investigate database cloud solutions. In my opinion, in our use case, NoSQL database will fit more than SQL ones. We could investigate solutions such as [Cloud Firestone](https://cloud.google.com/firestore/) or [Amazon DynamoDB](https://aws.amazon.com/dynamodb/).