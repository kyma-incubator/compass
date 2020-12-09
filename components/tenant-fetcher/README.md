# Tenant Fetcher

This component is introduced as a replacement for the preexisting tenant-fetcher CronJobs
which periodically fetch tenants from an external source.

Initially this component will only contain a simple HTTP server which is capable
of handling notification callbacks for the onboarding and decomissioning of a tenant
(this is new logic, not existent in the current CronJobs).

Eventully we will move all CronJobs to be part of this deployment and a Deployment
would be created for this new Tenant Fetcher which periodically fetches tenants
from the external source as opposed to spawning new CronJobs periodically.