# Requirements

## Storage

- Easy operations / no operations (Priority: super high)
- Ability to search Applications and Runtimes by labels (e.g. groups) (Priority: super high)
    - List all labels
    - List labels by key
    - List applications by label
- Cross-region replication (Priority: medium/low)
- The same storage type for the Runtime and Management Plane mode (Priority: high)

## Proof of Concept

Implement a simple app which:
- Creates apps with labels
- Uses Minio gateway with Amazon S3
- Queries apps by labels
