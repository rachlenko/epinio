// Package appchart collects the structures and functions that deal with epinio's app chart CR
package appchart

import (
	"context"
	"errors"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/helmchart"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// List returns a slice of all known app chart CRs.
func List(ctx context.Context, cluster *kubernetes.Cluster) (models.AppChartList, error) {
	client, err := cluster.ClientAppChart()
	if err != nil {
		return nil, err
	}

	list, err := client.Namespace(helmchart.Namespace()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	apps := make(models.AppChartList, 0, len(list.Items))

	for _, chart := range list.Items {
		copy := chart // Prevent memory aliasing warning
		appChart, err := toChart(&copy)
		if err != nil {
			return nil, err
		}
		apps = append(apps, appChart.AppChart)
	}

	return apps, nil
}

// Exists tests if the named app chart exists, or not.
func Exists(ctx context.Context, cluster *kubernetes.Cluster, name string) (bool, error) {
	_, err := Get(ctx, cluster, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Lookup returns the named app chart, or nil
func Lookup(ctx context.Context, cluster *kubernetes.Cluster, name string) (*models.AppChartFull, error) {
	chartCR, err := Get(ctx, cluster, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return toChart(chartCR)
}

// Get returns the app chart resource from the cluster.  This should be
// changed to return a typed application struct, like epinioappv1.AppChartSpec if
// needed in the future.
func Get(ctx context.Context, cluster *kubernetes.Cluster, name string) (*unstructured.Unstructured, error) {
	client, err := cluster.ClientAppChart()
	if err != nil {
		return nil, err
	}

	return client.Namespace(helmchart.Namespace()).Get(ctx, name, metav1.GetOptions{})
}

// toChart converts the unstructured app chart CR into the proper model
func toChart(chart *unstructured.Unstructured) (*models.AppChartFull, error) {

	name, _, err := unstructured.NestedString(chart.UnstructuredContent(), "metadata", "name")
	if err != nil {
		return nil, errors.New("chart should be string")
	}

	description, _, err := unstructured.NestedString(chart.UnstructuredContent(), "spec", "description")
	if err != nil {
		return nil, errors.New("description should be string")
	}

	short, _, err := unstructured.NestedString(chart.UnstructuredContent(), "spec", "shortDescription")
	if err != nil {
		return nil, errors.New("shortdescription should be string")
	}

	helmChart, _, err := unstructured.NestedString(chart.UnstructuredContent(), "spec", "helmChart")
	if err != nil {
		return nil, errors.New("helm chart should be string")
	}

	helmRepo, _, err := unstructured.NestedString(chart.UnstructuredContent(), "spec", "helmRepo")
	if err != nil {
		return nil, errors.New("helm repo should be string")
	}

	theValues, _, err := unstructured.NestedStringMap(chart.UnstructuredContent(), "spec", "values")
	if err != nil {
		return nil, errors.New("spec values should be string")
	}

	theSettings, _, err := unstructured.NestedMap(chart.UnstructuredContent(), "spec", "settings")
	if err != nil {
		return nil, errors.New("spec settings should be string")
	}

	settings := make(map[string]models.AppChartSetting)
	for key := range theSettings {
		fieldType, _, err := unstructured.NestedString(theSettings, key, "type")
		if err != nil {
			return nil, errors.New("settings type should be string")
		}
		fieldMin, _, err := unstructured.NestedString(theSettings, key, "minimum")
		if err != nil {
			return nil, errors.New("settings minimum should be string")
		}
		fieldMax, _, err := unstructured.NestedString(theSettings, key, "maximum")
		if err != nil {
			return nil, errors.New("settings maximum should be string")
		}
		fieldEnum, _, err := unstructured.NestedStringSlice(theSettings, key, "enum")
		if err != nil {
			return nil, errors.New("settings enum should be string slice")
		}

		settings[key] = models.AppChartSetting{
			Type:    fieldType,
			Minimum: fieldMin,
			Maximum: fieldMax,
			Enum:    fieldEnum,
		}
	}

	createdAt := chart.GetCreationTimestamp()

	return &models.AppChartFull{
		AppChart: models.AppChart{
			Meta: models.MetaLite{
				Name:      name,
				CreatedAt: createdAt,
			},
			Description:      description,
			ShortDescription: short,
			HelmChart:        helmChart,
			HelmRepo:         helmRepo,
			Settings:         settings,
		},
		Values: theValues,
	}, nil
}
