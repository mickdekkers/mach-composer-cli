### azure related
short_name              = "{{ component.short_name }}"
name_prefix             = local.name_prefix
subscription_id         = local.subscription_id
tenant_id               = local.tenant_id
service_object_ids      = local.service_object_ids
region                  = local.region
resource_group_name     = azurerm_resource_group.main.name
resource_group_location = azurerm_resource_group.main.location
app_service_plan_id     = azurerm_app_service_plan.functionapps.id
tags                    = local.tags
{% if site.azure.alert_group %}
monitor_action_group_id = azurerm_monitor_action_group.alert_action_group.id
{% endif %}