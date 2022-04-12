package namespace

import (
	"context"
	"sync"
	"time"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/internal/configurations"
	"github.com/epinio/epinio/internal/namespaces"
	"github.com/epinio/epinio/internal/services"
	apierror "github.com/epinio/epinio/pkg/api/core/v1/errors"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
	"github.com/gin-gonic/gin"

	ants "github.com/panjf2000/ants/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Delete handles the API endpoint /namespaces/:namespace (DELETE).
// It destroys the namespace specified by its name.
// This includes all the applications and configurations in it.
func (oc Controller) Delete(c *gin.Context) apierror.APIErrors {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	cluster, err := kubernetes.GetCluster(ctx)
	if err != nil {
		return apierror.InternalError(err)
	}

	exists, err := namespaces.Exists(ctx, cluster, namespace)
	if err != nil {
		return apierror.InternalError(err)
	}
	if !exists {
		return apierror.NamespaceIsNotKnown(namespace)
	}

	err = deleteApps(ctx, cluster, namespace)
	if err != nil {
		return apierror.InternalError(err)
	}

	err = deleteServices(ctx, cluster, namespace)
	if err != nil {
		return apierror.InternalError(err)
	}

	configurationList, err := configurations.List(ctx, cluster, namespace)
	if err != nil {
		return apierror.InternalError(err)
	}

	for _, configuration := range configurationList {
		err = configuration.Delete(ctx)
		if err != nil && !apierrors.IsNotFound(err) {
			return apierror.InternalError(err)
		}
	}

	// Deleting the namespace here. That will automatically delete the application resources.
	err = namespaces.Delete(ctx, cluster, namespace)
	if err != nil {
		return apierror.InternalError(err)
	}

	response.OK(c)
	return nil
}

// deleteApps removes the application and its resources
func deleteApps(ctx context.Context, cluster *kubernetes.Cluster, namespace string) error {
	appRefs, err := application.ListAppRefs(ctx, cluster, namespace)
	if err != nil {
		return err
	}

	const maxConcurrent = 100
	errChan := make(chan error)

	var wg, errWg sync.WaitGroup
	var loopErr error

	errWg.Add(1)
	go func() {
		for err := range errChan {
			loopErr = err
			break
		}
		errWg.Done()
	}()

	p, err := ants.NewPoolWithFunc(maxConcurrent, func(i interface{}) {
		err := application.Delete(ctx, cluster, i.(models.AppRef))
		if err != nil {
			errChan <- err
		}
		wg.Done()
	}, ants.WithExpiryDuration(10*time.Second))
	if err != nil {
		return err
	}

	for _, appRef := range appRefs {
		wg.Add(1)
		err = p.Invoke(appRef)
		if err != nil {
			errChan <- err
		}
	}
	defer p.Release()

	wg.Wait()
	close(errChan)
	errWg.Wait()

	return loopErr
}

// deleteServices removes all provisioned services when a Namespace is deleted
func deleteServices(ctx context.Context, cluster *kubernetes.Cluster, namespace string) error {
	kubeServiceClient, err := services.NewKubernetesServiceClient(cluster)
	if err != nil {
		return apierror.InternalError(err)
	}

	return kubeServiceClient.DeleteAll(ctx, namespace)
}
