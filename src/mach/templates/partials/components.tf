{% if site.azure and site.components %}
resource "azurerm_app_service_plan" "functionapps" {
  name                = format("%s-plan", local.name_prefix)
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  kind                = "FunctionApp"
  reserved            = true

  sku {
    tier = "Dynamic"
    size = "Y1"
  }

  tags = local.tags
}

{% for component in site.components %}
{% include 'partials/component.tf' %}
{% endfor %}
{% endif %}