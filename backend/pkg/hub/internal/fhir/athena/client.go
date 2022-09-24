package athena

import (
	"context"
	"fmt"
	"github.com/fastenhealth/fastenhealth-onprem/backend/pkg/config"
	"github.com/fastenhealth/fastenhealth-onprem/backend/pkg/database"
	"github.com/fastenhealth/fastenhealth-onprem/backend/pkg/hub/internal/fhir/base"
	"github.com/fastenhealth/fastenhealth-onprem/backend/pkg/models"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AthenaClient struct {
	*base.FHIR401Client
}

func NewClient(ctx context.Context, appConfig config.Interface, globalLogger logrus.FieldLogger, source models.Source, testHttpClient ...*http.Client) (base.Client, *models.Source, error) {
	baseClient, updatedSource, err := base.NewFHIR401Client(ctx, appConfig, globalLogger, source, testHttpClient...)
	return AthenaClient{
		baseClient,
	}, updatedSource, err
}

func (c AthenaClient) SyncAll(db database.DatabaseRepository) error {

	supportedResources := []string{
		"AllergyIntolerance",
		//"Binary",
		"CarePlan",
		"CareTeam",
		"Condition",
		"Device",
		"DiagnosticReport",
		"DocumentReference",
		"Encounter",
		"Goal",
		"Immunization",
		//"Location",
		//"Medication",
		//"MedicationRequest",
		"Observation",
		//"Organization",
		//"Patient",
		//"Practitioner",
		"Procedure",
		//"Provenance",
	}
	for _, resourceType := range supportedResources {
		bundle, err := c.GetResourceBundle(fmt.Sprintf("%s?patient=%s", resourceType, c.Source.PatientId))
		if err != nil {
			return err
		}
		wrappedResourceModels, err := c.ProcessBundle(bundle)
		if err != nil {
			c.Logger.Infof("An error occurred while processing %s bundle %s", resourceType, c.Source.PatientId)
			return err
		}
		//todo, create the resources in dependency order

		for _, apiModel := range wrappedResourceModels {
			err = db.UpsertResource(context.Background(), apiModel)
			if err != nil {
				return err
			}
		}
	}
	return nil

}