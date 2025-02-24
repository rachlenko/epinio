package application

// Validate the custom chart values saved in the application CR against the declarations made by the
// app chart referenced by the application CR.

import (
	"github.com/gin-gonic/gin"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/appchart"
	"github.com/epinio/epinio/internal/application"
	apierror "github.com/epinio/epinio/pkg/api/core/v1/errors"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
)

// ValidateChartValues handles the API endpoint /namespaces/:namespace/applications/:app/validate-cv
// Given application by name, and namespace the custom chart values are checked against the
// declarations in the referenced appchart.
func (hc Controller) ValidateChartValues(c *gin.Context) apierror.APIErrors {
	ctx := c.Request.Context()

	namespace := c.Param("namespace")
	appName := c.Param("app")

	cluster, err := kubernetes.GetCluster(ctx)
	if err != nil {
		return apierror.InternalError(err)
	}

	appRef := models.NewAppRef(appName, namespace)
	exists, err := application.Exists(ctx, cluster, appRef)
	if err != nil {
		return apierror.InternalError(err)
	}

	if !exists {
		return apierror.AppIsNotKnown(appName)
	}

	app, err := application.Lookup(ctx, cluster, namespace, appName)
	if err != nil {
		return apierror.InternalError(err)
	}

	appChart, err := appchart.Lookup(ctx, cluster, app.Configuration.AppChart)
	if err != nil {
		return apierror.InternalError(err)
	}

	if appChart == nil {
		return apierror.AppChartIsNotKnown(app.Configuration.AppChart)
	}

	issues := application.ValidateCV(app.Configuration.Settings, appChart.Settings)
	if issues != nil {
		// Treating all validation failures as internal errors.
		// I can't find something better at the moment.

		var apiIssues []apierror.APIError
		for _, err := range issues {
			apiIssues = append(apiIssues, apierror.NewBadRequestError(err.Error()))
		}

		return apierror.NewMultiError(apiIssues)
	}

	// Return the id of the new blob
	response.OK(c)
	return nil
}
