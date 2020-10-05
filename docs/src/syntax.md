# Configuration syntax

A configuration file can contain several sites with all different configurations and all using a different mix of re-usable serverless microservice components.

It is common to have a single configuration file per environment since they usually share the same general configurations.

The configuration file has the following structure:

- **[general_config](#general_config)**
    - **[environment](#general_config)**
    - **[terraform_config](#terraform_config)**
    - **[cloud](#general_config)**
    - [azure](#azure)
    - [sentry](#sentry)
- **[sites](#sites)**
    - **[identifier](#sites)**
    - [commercetools](#commercetools)
    - [azure](#azure_1)
    - [aws](#aws)
    - [stores](#stores)
    - [components](#component-configurations)
- [components](#components)

## general_config
All 'shared' configuration that applies to all sites.

- **`environment`** - (Required) [environment](#environment) Identifier for the environment. Must be one of `development`, `test` or `production`.  
Is used to set the `environment` variable of any [Terraform component](./components.md#terraform-component)
- **`terraform_config`** - (Required) [terraform_config](#terraform_config) block
- `cloud` - Either `azure` or `aws`
- `azure` - [Azure](#azure) block
- `sentry` - [Sentry](#sentry) block


### terraform_config
Configuration of the Terraform state backend.  
This can be any of:

- Azure Blob Container
- AWS S3

#### Azure storage
An Azure state backend can be defined as:

```
terraform_config:
  azure_remote_state:
    resource_group_name: <your resource group>
    storage_account_name: <storage account name>
    container_name: <container name>
    state_folder: <state folder>
```

!!! tip ""
    A good convention is to give the state_folder the same name as [`environment`](#environment)

#### AWS S3
An AWS S3 state backend can be defined as:

```
terraform_config:
  aws_remote_state:
    bucket: mach-statefiles
    key_prefix: test-statefiles
    role_arn: arn:aws:iam::1234567890:role/deploy
```

- **`bucket`** - (Required) S3 bucket name
- **`key_prefix`** - (Required) Key prefix for each individual Terraform state
- **`role_arn`** - (Required) Role ARN to access S3 bucket with
- `lock_table` - DynamoDB lock table

### sentry
Defines a Sentry configuration.

Example:
```
sentry:
  dsn: https://LhNrqROZRIl2c5ciidkN82DObJfgtiLd@sentry.io/123456
```

When defined, a `sentry_dsn` variable is passed on to all Terraform components.

!!! tip ""
    For more information about data exposed to Terraform modules, see the [Terraform component](./components.md#terraform-component) docs

### azure

General Azure settings. Values can be overwritten [per site](#azure_1).

Example:
```
azure:
  tenant_id: f2e03b8b-fe10-4fbc-9f5c-76dad9ac52e2
  subscription_id: a5b51c09-a2da-45b8-918a-67cf42456ab3
  region: westeurope
  resources_prefix: my-
  service_object_ids:
    gitlab-sp: d1114ea6-88f9-45b2-9de4-031291090380 # gitlab-sp
    developers: 3d280212-934f-4d32-876d-1b541a7697ba # developers tst group
```

- **`tenant_id`** - (Required)
- **`subscription_id`** - (Required)
- **`region`** - (Required)
- `resources_prefix` - 
- `front_door` - [Front-door](#frontdoor) settings
- `service_object_ids` - Map of service objects IDs that should have access to the components KeyVault.

#### front_door

Example:
```
front_door:
  resource_group_name: my-shared-rg
  dns_zone: my-services-domain.net
  ssl_key_vault_name: mysharedwekvcdn
  ssl_key_vault_secret_name: wildcard-my-services-domain-net
  ssl_key_vault_secret_version: IOlB8XmYLH1keYcpkcji23sp
```

- **`resource_group_name`** - (Required)
- **`dns_zone`** - (Required)
- **`ssl_key_vault_name`** - (Required)
- **`ssl_key_vault_secret_name`** - (Required)
- **`ssl_key_vault_secret_version`** - (Required)


## sites
All site definitions.


- **`identifier`** - (Required)  
  Unique identifier for this site.  
  Will be used for the Terraform state and naming all cloud resources.
- `commercetools` - [commercetools configuration](#commercetools) block
- `azure` - [Azure](#azure_1) settings
- `aws` - [AWS](#aws) settings
- `components` - [Component configurations](#component-configurations)

### commercetools

commercetools configuration.

Example:

```
commercetools:
  project_key: my-site-tst
  client_id: T9J5g5bJe-VV8aVvN5Q
  client_secret: FIo3PGHJDThCM17wok_irLakRzCA
  scopes: manage_api_clients:my-site-tst manage_project:my-site-tst view_api_clients:my-site-tst
  languages:
    - en-GB
    - nl-NL
  currencies:
    - GBP
    - EUR
  countries:
    - GB
    - NL
```

- **`project_key`** - (Required) commercetools project key
- **`client_id`** - (Required) API client ID
- **`client_secret`** - (Required) API client secret
- **`scopes`** - (Required) Required scopes for given API client ID.
- `token_url` - Defaults to `https://auth.europe-west1.gcp.commercetools.com`
- `api_url` - Defaults to `https://api.europe-west1.gcp.commercetools.com`
- `currencies` - List of three-digit currency codes as per ISO 4217
- `languages` - List of IETF language tag
- `countries` - List of two-digit country codes as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
- `messages_enabled` - When false the creation of messages is disabled.  
  Defaults to True
- `channels` - List of [channel definitions](#channels)
- `taxes` - List of [tax definitions](#tax)
- `stores` - List of [store definitions](#store) if multiple (store) contexts are going to be used.
- `create_frontend_credentials` - Defines if frontend API credentials must be created
  Defaults to `true`

#### channels

Example
```
channels:
  - key: INV
    roles:
      - InventorySupply
    name:
      en-GB: Inventory
    description:
      en-GB: Our main inventory channel
  - key: DIST
    roles:
      - ProductDistribution
    name:
      en-GB: Distribution
    description:
      en-GB: Our main distribution channel
```

- **`key`** - (Required) 
- **`roles`** - (Required) List of [channel roles](https://docs.commercetools.com/http-api-projects-channels#channelroleenum).  
    Can be one of `InventorySupply`, `ProductDistribution`, `OrderExport`, `OrderImport` or `Primary`
- `name` - Name of the channel. Localized string [^1]
- `description` - Description of the channel. Localized string [^1]

#### taxes

Defines tax rates for various countries.

- **`country`** - (Required) A two-digit country code as per [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
- **`amount`** - (Required) Number Percentage in the range of [0..1]
- **`name`** - (Required) Tax rate name

#### stores
Defines [commercetools stores](https://docs.commercetools.com/http-api-projects-stores).

Example:
```
stores:
  - name:
      en-GB: my store
    key: mystore
    distribution_channels:
      - DIST
```

- **`name`** - (Required) Name of the store. Localized string [^1]
- **`key`** - (Required) Store key
- `languages` - (Required)
- `distribution_channels` - (Required)


### azure
Site-specific Azure settings.  
Can overwrite any value from the generic [Azure settings](#azure):

- `tenant_id`
- `service_object_ids`
- `front_door`
- `subscription_id`
- `region`

And adds the following exta attributes:

- `alert_group` - List of [Alert groups](#alert_group)

#### alert_group
Example:

```
alert_group:
  name: critical
  alert_emails:
    - alerting@example.com
  logic_app: my-shared-we-rg.my-shared-we-alerts-slack
  webhook_url: https://example.com/api/alert-me/
```

- `name` - (Required) The name of the alert group
- `alert_emails` - Hook alert group to these email addresses
- `webhook_url` - Hooks alert group to a webhook
- `logic_app` - Reference to a Logic App the alert group needs to be connected to.  
  Format is `<resource_group_name>.<logic_app_name>`

### aws
Site-specific AWS settings.

Example:

```
aws:
  account_id: 1234567890
  region: eu-west-1
  deploy_role: deploy
  api_gateway: main_gateway
  extra_providers:
    - name: email
      region: eu-west-1
```

- **`account_id`** - (Required) AWS account ID for this site
- **`region`** - AWS region to deploy site in
- **`deploy_role`** - (Required) The [IAM role](./prerequisites.md#deploy-iam-role) needed for deployment
- `api_gateway` - Name of the main API gateway
- `extra_providers`


### component configurations

- **`name`** - (Required) Reference to a [component](#component) definition
- `variables` - Environment variables for this components runtime
- `secrets` - Environment variables for this component that should be stored in a encrypted key-value store
- `health_check_path` - Defines a custom healthcheck path.  
  Overwrites the default `health_check_path` defined in the component definition

## components

Component definitions.  
These components are used and configured separately [per site](#component-configurations).

Example:

```
components:
  - name: api-extensions
    short_name: apiexts
    source: git::ssh://git@git.labdigital.nl/mach-components/api-extensions-component.git//terraform
    version: 3b8ab91
    has_public_api: true
  - name: ct-products-types
    source: git::ssh://git@git.labdigital.nl/mach-components/ct-product-types.git//terraform
    version: 1.4.0
    is_software_component: false
```

- `version` - A Git commit hash or tag
- `source` - Source definition of the terraform module
- `short_name` - Short name to be used in cloud resources. Should be at most 10 characters to avoid running into Resource naming limits.  
  Defaults to the given components `name`
- `is_software_component` - Defines if this is a 'software component' meaing; contains any serverless function.  
  Defaults to `true`
- `has_public_api` - Defines if the serverless function should be exposed publically.  
  Will create proper Frontdoor routing when `true`.
- `health_check_path` - Defines a custom healthcheck path.  
  Defaults to `/<name>/healthchecks`

[^1]: commercetools uses [Localized strings](https://docs.commercetools.com/http-api-types#localizedstring) to be able to define strings in mulitple languages.  
Whenever a localized string needs to be defined, this can be done in the following format:
```
some-string:
  - en-GB: My value
  - nl-NL:  Mijn waarde
```