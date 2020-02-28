package instana_test

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"

	. "github.com/gessnerfl/terraform-provider-instana/instana"
	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/testutils"
	"github.com/gessnerfl/terraform-provider-instana/utils"
)

var testCustomEventSpecificationWithSystemRuleProviders = map[string]terraform.ResourceProvider{
	"instana": Provider(),
}

const resourceCustomEventSpecificationWithSystemRuleDefinitionTemplate = `
provider "instana" {
  api_token = "test-token"
  endpoint = "localhost:{{PORT}}"
}

resource "instana_custom_event_spec_system_rule" "example" {
  name = "name"
  query = "query"
  enabled = true
  triggering = true
  description = "description"
  expiration_time = 60000
  rule_severity = "warning"
  rule_system_rule_id = "system-rule-id"
  downstream_integration_ids = [ "integration-id-1", "integration-id-2" ]
  downstream_broadcast_to_all_alerting_configs = true
}
`

const (
	customSystemEventApiPath                             = restapi.CustomEventSpecificationResourcePath + "/{id}"
	testCustomEventSpecificationWithSystemRuleDefinition = "instana_custom_event_spec_system_rule.example"

	customSystemEventID                       = "custom-system-event-id"
	customSystemEventName                     = "name"
	customSystemEventQuery                    = "query"
	customSystemEventExpirationTime           = 60000
	customSystemEventDescription              = "description"
	customSystemEventRuleSystemRuleId         = "system-rule-id"
	customSystemEventDownStringIntegrationId1 = "integration-id-1"
	customSystemEventDownStringIntegrationId2 = "integration-id-2"

	customSystemEventMessageNotAValidSeverity           = "not a valid severity"
	customSystemEventTestMessageExpectedInvalidSeverity = "Expected to get error that the provided severity is not valid"

	constSystemEventContentType = "Content-Type"
)

var customSystemEventRuleSeverity = restapi.SeverityWarning.GetTerraformRepresentation()

func TestCRUDOfCreateResourceCustomEventSpecificationWithSystemdRuleResourceWithMockServer(t *testing.T) {
	testutils.DeactivateTLSServerCertificateVerification()
	httpServer := testutils.NewTestHTTPServer()
	httpServer.AddRoute(http.MethodPut, customSystemEventApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodDelete, customSystemEventApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodGet, customSystemEventApiPath, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		//the integration id order is changed by intention to ensure set behaviour
		json := strings.ReplaceAll(`
		{
			"id" : "{{id}}",
			"name" : "name",
			"entityType" : "entity_type",
			"query" : "query",
			"enabled" : true,
			"triggering" : true,
			"description" : "description",
			"expirationTime" : 60000,
			"rules" : [ { "ruleType" : "system", "severity" : 5, "systemRuleId" : "system-rule-id" } ],
			"downstream" : {
				"integrationIds" : ["integration-id-2", "integration-id-1"],
				"broadcastToAllAlertingConfigs" : true
			}
		}
		`, "{{id}}", vars["id"])
		w.Header().Set(constSystemEventContentType, r.Header.Get(constSystemEventContentType))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	})
	httpServer.Start()
	defer httpServer.Close()

	resourceCustomEventSpecificationWithSystemRuleDefinition := strings.ReplaceAll(resourceCustomEventSpecificationWithSystemRuleDefinitionTemplate, "{{PORT}}", strconv.Itoa(httpServer.GetPort()))

	hashFunctionDownstreamIntegrationId := schema.HashSchema(CustomEventSpecificationSchemaDownstreamIntegrationIds.Elem.(*schema.Schema))
	resource.UnitTest(t, resource.TestCase{
		Providers: testCustomEventSpecificationWithSystemRuleProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: resourceCustomEventSpecificationWithSystemRuleDefinition,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testCustomEventSpecificationWithSystemRuleDefinition, "id"),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldName, customSystemEventName),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldEntityType, SystemRuleEntityType),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldQuery, customSystemEventQuery),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldTriggering, "true"),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldDescription, customSystemEventDescription),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldExpirationTime, strconv.Itoa(customSystemEventExpirationTime)),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationFieldEnabled, "true"),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, fmt.Sprintf("%s.%d", CustomEventSpecificationDownstreamIntegrationIds, hashFunctionDownstreamIntegrationId(customEventSpecificationWithThresholdRuleDownstreamIntegrationId1)), customSystemEventDownStringIntegrationId1),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, fmt.Sprintf("%s.%d", CustomEventSpecificationDownstreamIntegrationIds, hashFunctionDownstreamIntegrationId(customEventSpecificationWithThresholdRuleDownstreamIntegrationId2)), customSystemEventDownStringIntegrationId2),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs, "true"),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, CustomEventSpecificationRuleSeverity, customSystemEventRuleSeverity),
					resource.TestCheckResourceAttr(testCustomEventSpecificationWithSystemRuleDefinition, SystemRuleSpecificationSystemRuleID, customSystemEventRuleSystemRuleId),
				),
			},
		},
	})
}

func TestCustomEventSpecificationWithSystemRuleSchemaDefinitionIsValid(t *testing.T) {
	schema := NewCustomEventSpecificationWithSystemRuleResourceHandle().Schema

	schemaAssert := testutils.NewTerraformSchemaAssert(schema, t)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(CustomEventSpecificationFieldName)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(CustomEventSpecificationFieldFullName)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(CustomEventSpecificationFieldEntityType)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(CustomEventSpecificationFieldQuery)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationFieldTriggering, false)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(CustomEventSpecificationFieldDescription)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeInt(CustomEventSpecificationFieldExpirationTime)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationFieldEnabled, true)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeSetOfStrings(CustomEventSpecificationDownstreamIntegrationIds)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs, true)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(CustomEventSpecificationRuleSeverity)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(SystemRuleSpecificationSystemRuleID)
}

func TestCustomEventSpecificationWithSystemRuleResourceShouldHaveSchemaVersionOne(t *testing.T) {
	assert.Equal(t, 2, NewCustomEventSpecificationWithSystemRuleResourceHandle().SchemaVersion)
}

func TestCustomEventSpecificationWithSystemRuleShouldHaveOneStateUpgraderForVersionZero(t *testing.T) {
	resourceHandler := NewCustomEventSpecificationWithSystemRuleResourceHandle()

	assert.Equal(t, 2, len(resourceHandler.StateUpgraders))
	assert.Equal(t, 0, resourceHandler.StateUpgraders[0].Version)
	assert.Equal(t, 1, resourceHandler.StateUpgraders[1].Version)
}

func TestShouldMigrateCustomEventSpecificationWithSystemRuleStateAndAddFullNameWithSameValueAsNameWhenMigratingFromVersion0To1(t *testing.T) {
	name := "Test Name"
	rawData := make(map[string]interface{})
	rawData[CustomEventSpecificationFieldName] = name
	meta := "dummy"

	result, err := NewCustomEventSpecificationWithSystemRuleResourceHandle().StateUpgraders[0].Upgrade(rawData, meta)

	assert.Nil(t, err)
	assert.Equal(t, name, result[CustomEventSpecificationFieldFullName])
}

func TestShouldMigrateEmptyCustomEventSpecificationWithSystemRuleStateFromVersion0To1(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"

	result, err := NewCustomEventSpecificationWithSystemRuleResourceHandle().StateUpgraders[0].Upgrade(rawData, meta)

	assert.Nil(t, err)
	assert.Nil(t, result[CustomEventSpecificationFieldFullName])
}

func TestShouldMigrateDownstreamIntegrationIdsOfCustomEventSpecificationWithSystemRuleFromListToSetWhenMigrationFromVersion1To2(t *testing.T) {
	integrationIds := []interface{}{"id1", "id2"}
	rawData := make(map[string]interface{})
	rawData[CustomEventSpecificationDownstreamIntegrationIds] = integrationIds
	meta := "dummy"

	result, err := NewCustomEventSpecificationWithSystemRuleResourceHandle().StateUpgraders[1].Upgrade(rawData, meta)

	assert.Nil(t, err)
	assert.IsType(t, schema.NewSet(schema.HashString, integrationIds), result[CustomEventSpecificationDownstreamIntegrationIds])
	assert.Len(t, result[CustomEventSpecificationDownstreamIntegrationIds].(*schema.Set).List(), 2)
	assert.Contains(t, result[CustomEventSpecificationDownstreamIntegrationIds].(*schema.Set).List(), "id1")
	assert.Contains(t, result[CustomEventSpecificationDownstreamIntegrationIds].(*schema.Set).List(), "id2")
}

func TestShouldMigrateEmptyCustomEventSpecificationWithSystemRuleStateFromVersion1To2(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"

	result, err := NewCustomEventSpecificationWithSystemRuleResourceHandle().StateUpgraders[1].Upgrade(rawData, meta)

	assert.Nil(t, err)
	assert.Nil(t, result[CustomEventSpecificationDownstreamIntegrationIds])
}

func TestShouldReturnCorrectResourceNameForCustomEventSpecificationWithSystemRuleResource(t *testing.T) {
	name := NewCustomEventSpecificationWithSystemRuleResourceHandle().ResourceName

	assert.Equal(t, name, "instana_custom_event_spec_system_rule")
}

func TestShouldUpdateCustomEventSpecificationWithSystemRuleTerraformStateFromApiObject(t *testing.T) {
	description := customSystemEventDescription
	expirationTime := customSystemEventExpirationTime
	query := customSystemEventQuery
	spec := restapi.CustomEventSpecification{
		ID:             customSystemEventID,
		Name:           customSystemEventName,
		EntityType:     SystemRuleEntityType,
		Query:          &query,
		Description:    &description,
		ExpirationTime: &expirationTime,
		Triggering:     true,
		Enabled:        true,
		Rules: []restapi.RuleSpecification{
			restapi.NewSystemRuleSpecification(customSystemEventRuleSystemRuleId, restapi.SeverityWarning.GetAPIRepresentation()),
		},
		Downstream: &restapi.EventSpecificationDownstream{
			IntegrationIds:                []string{customSystemEventDownStringIntegrationId1, customSystemEventDownStringIntegrationId2},
			BroadcastToAllAlertingConfigs: true,
		},
	}

	testHelper := NewTestHelper(t)
	sut := NewCustomEventSpecificationWithSystemRuleResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, spec)

	assert.Nil(t, err)
	assert.Equal(t, customSystemEventID, resourceData.Id())
	assert.Equal(t, customSystemEventName, resourceData.Get(CustomEventSpecificationFieldFullName))
	assert.Equal(t, SystemRuleEntityType, resourceData.Get(CustomEventSpecificationFieldEntityType))
	assert.Equal(t, customSystemEventQuery, resourceData.Get(CustomEventSpecificationFieldQuery))
	assert.Equal(t, description, resourceData.Get(CustomEventSpecificationFieldDescription))
	assert.True(t, resourceData.Get(CustomEventSpecificationFieldTriggering).(bool))
	assert.True(t, resourceData.Get(CustomEventSpecificationFieldEnabled).(bool))

	assert.Equal(t, customSystemEventRuleSystemRuleId, resourceData.Get(SystemRuleSpecificationSystemRuleID))
	assert.Equal(t, restapi.SeverityWarning.GetTerraformRepresentation(), resourceData.Get(CustomEventSpecificationRuleSeverity))

	assert.True(t, resourceData.Get(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs).(bool))
	assert.NotNil(t, resourceData.Get(CustomEventSpecificationDownstreamIntegrationIds))
	integrationIds := resourceData.Get(CustomEventSpecificationDownstreamIntegrationIds).(*schema.Set)
	assert.Len(t, integrationIds.List(), 2)
	assert.Contains(t, integrationIds.List(), customSystemEventDownStringIntegrationId1)
	assert.Contains(t, integrationIds.List(), customSystemEventDownStringIntegrationId2)
}

func TestShouldSuccessfullyConvertCustomEventSpecificationWithSystemRuleStateToDataModel(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewCustomEventSpecificationWithSystemRuleResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)

	resourceData.SetId(customSystemEventID)
	resourceData.Set(CustomEventSpecificationFieldFullName, customSystemEventName)
	resourceData.Set(CustomEventSpecificationFieldEntityType, SystemRuleEntityType)
	resourceData.Set(CustomEventSpecificationFieldQuery, customSystemEventQuery)
	resourceData.Set(CustomEventSpecificationFieldTriggering, true)
	resourceData.Set(CustomEventSpecificationFieldDescription, customSystemEventDescription)
	resourceData.Set(CustomEventSpecificationFieldExpirationTime, customSystemEventExpirationTime)
	resourceData.Set(CustomEventSpecificationFieldEnabled, true)
	integrationIds := make([]interface{}, 2)
	integrationIds[0] = customSystemEventDownStringIntegrationId1
	integrationIds[1] = customSystemEventDownStringIntegrationId2
	resourceData.Set(CustomEventSpecificationDownstreamIntegrationIds, integrationIds)
	resourceData.Set(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs, true)
	resourceData.Set(CustomEventSpecificationRuleSeverity, customSystemEventRuleSeverity)
	resourceData.Set(SystemRuleSpecificationSystemRuleID, customSystemEventRuleSystemRuleId)

	result, err := resourceHandle.MapStateToDataObject(resourceData, utils.NewResourceNameFormatter("prefix ", " suffix"))

	assert.Nil(t, err)
	assert.IsType(t, restapi.CustomEventSpecification{}, result)
	customEventSpec := result.(restapi.CustomEventSpecification)
	assert.Equal(t, customSystemEventID, customEventSpec.GetID())
	assert.Equal(t, customSystemEventName, customEventSpec.Name)
	assert.Equal(t, SystemRuleEntityType, customEventSpec.EntityType)
	assert.Equal(t, customSystemEventQuery, *customEventSpec.Query)
	assert.Equal(t, customSystemEventDescription, *customEventSpec.Description)
	assert.Equal(t, customSystemEventExpirationTime, *customEventSpec.ExpirationTime)
	assert.True(t, customEventSpec.Triggering)
	assert.True(t, customEventSpec.Enabled)

	assert.Equal(t, 1, len(customEventSpec.Rules))
	assert.Equal(t, customSystemEventRuleSystemRuleId, *customEventSpec.Rules[0].SystemRuleID)
	assert.Equal(t, restapi.SeverityWarning.GetAPIRepresentation(), customEventSpec.Rules[0].Severity)

	assert.True(t, customEventSpec.Downstream.BroadcastToAllAlertingConfigs)
	assert.Len(t, customEventSpec.Downstream.IntegrationIds, 2)
	assert.Contains(t, customEventSpec.Downstream.IntegrationIds, customSystemEventDownStringIntegrationId1)
	assert.Contains(t, customEventSpec.Downstream.IntegrationIds, customSystemEventDownStringIntegrationId2)
}

func TestShouldFailToConvertCustomEventSpecificationWithSystemRuleStateToDataModelWhenSeverityIsNotValid(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewCustomEventSpecificationWithSystemRuleResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.Set(CustomEventSpecificationRuleSeverity, "INVALID")

	_, err := resourceHandle.MapStateToDataObject(resourceData, utils.NewResourceNameFormatter("prefix ", " suffix"))

	assert.NotNil(t, err)
}
