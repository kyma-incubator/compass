# Requirements

## Storage

- Easy operations / no operations (Priority: super high)
- Ability to search Applications and Runtimes by labels (e.g. groups) (Priority: super high)
    - List all labels
    - List labels by key
    - List applications by label
- Cross-region replication (Priority: medium/low)
- Same storage type for Runtime and Management Plane mode (Priority: high)

## Proof of Concept

Implement a simple app which:
- Create apps with labels
- Use Minio gateway and directly for S3
- Query apps by labels
