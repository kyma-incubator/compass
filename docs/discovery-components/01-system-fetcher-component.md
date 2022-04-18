# Application Discovery

## System Fetcher

Compass finds new applications (systems) in the following ways:
- **Manually** - When a client registers an application via the Director GraphQL API.
- **Automatically** - When the System Fetcher component is configured to fetch applications from an external registry.

The System Fetcher is modeled like a Kubernetes CronJob, which runs periodically and synchronizes the applications created by it in Compass with the applications available externally.

### System Creation
System Fetcher creates applications via application templates. It uses a configuration mapping between an external system property and a given application template name. Then, it creates the application from the template with the externally-provided properties, which also map to Compass properties (for example, name, base URL, etc.).
