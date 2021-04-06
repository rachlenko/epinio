package deployments

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
	"github.com/suse/carrier/helpers"
	"github.com/suse/carrier/kubernetes"
	"github.com/suse/carrier/termui"
	"github.com/suse/carrier/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Carrier struct {
	Debug   bool
	Timeout time.Duration
}

const (
	CarrierDeploymentID  = "carrier"
	carrierBinaryPVCYaml = "carrier/binary-pvc.yaml"
	carrierCopierPodYaml = "carrier/copier-pod.yaml"
	carrierServerYaml    = "carrier/server.yaml"
)

func (k *Carrier) ID() string {
	return CarrierDeploymentID
}

func (k *Carrier) Backup(c *kubernetes.Cluster, ui *termui.UI, d string) error {
	return nil
}

func (k *Carrier) Restore(c *kubernetes.Cluster, ui *termui.UI, d string) error {
	return nil
}

func (k Carrier) Describe() string {
	return emoji.Sprintf(":cloud:Carrier version: %s\n", version.Version)
}

// Delete removes Carrier from kubernetes cluster
func (k Carrier) Delete(c *kubernetes.Cluster, ui *termui.UI) error {
	ui.Note().KeeplineUnder(1).Msg("Removing Carrier...")

	existsAndOwned, err := c.NamespaceExistsAndOwned(CarrierDeploymentID)
	if err != nil {
		return errors.Wrapf(err, "failed to check if namespace '%s' is owned or not", CarrierDeploymentID)
	}
	if !existsAndOwned {
		ui.Exclamation().Msg("Skipping Carrier because namespace either doesn't exist or not owned by Carrier")
		return nil
	}

	message := "Deleting Carrier namespace " + CarrierDeploymentID
	_, err = helpers.WaitForCommandCompletion(ui, message,
		func() (string, error) {
			return "", c.DeleteNamespace(CarrierDeploymentID)
		},
	)
	if err != nil {
		return errors.Wrapf(err, "Failed deleting namespace %s", CarrierDeploymentID)
	}

	ui.Success().Msg("Carrier removed")

	return nil
}

func (k Carrier) apply(c *kubernetes.Cluster, ui *termui.UI, options kubernetes.InstallationOptions, upgrade bool) error {
	if err := k.createCarrierNamespace(c, ui); err != nil {
		return err
	}

	if out, err := applyCarrierServerYaml(c, ui); err != nil {
		return errors.Wrap(err, out)
	}

	// NOTE: Set CARRIER_DONT_WAIT_FOR_DEPLOYMENT when doing development to let
	// the installation continue. You can use the `make patch-carrier-deployment` target
	// later to fix the failing deployment.
	// TODO: Add link to docs
	if os.Getenv("CARRIER_DONT_WAIT_FOR_DEPLOYMENT") == "" {
		if err := c.WaitUntilPodBySelectorExist(ui, CarrierDeploymentID, "app.kubernetes.io/name=carrier-server", k.Timeout); err != nil {
			return errors.Wrap(err, "failed waiting Carrier carrier-server deployment to exist")
		}
		if err := c.WaitForPodBySelectorRunning(ui, CarrierDeploymentID, "app.kubernetes.io/name=carrier-server", k.Timeout); err != nil {
			return errors.Wrap(err, "failed waiting Carrier carrier-server deployment to be running")
		}
	}

	ui.Success().Msg("Carrier deployed")

	return nil
}

func (k Carrier) GetVersion() string {
	return version.Version
}

func (k Carrier) Deploy(c *kubernetes.Cluster, ui *termui.UI, options kubernetes.InstallationOptions) error {

	_, err := c.Kubectl.CoreV1().Namespaces().Get(
		context.Background(),
		CarrierDeploymentID,
		metav1.GetOptions{},
	)
	if err == nil {
		return errors.New("Namespace " + CarrierDeploymentID + " present already")
	}

	ui.Note().KeeplineUnder(1).Msg("Deploying Carrier...")

	err = k.apply(c, ui, options, false)
	if err != nil {
		return err
	}

	return nil
}

func (k Carrier) Upgrade(c *kubernetes.Cluster, ui *termui.UI, options kubernetes.InstallationOptions) error {
	_, err := c.Kubectl.CoreV1().Namespaces().Get(
		context.Background(),
		CarrierDeploymentID,
		metav1.GetOptions{},
	)
	if err != nil {
		return errors.New("Namespace " + CarrierDeploymentID + " not present")
	}

	ui.Note().Msg("Upgrading Carrier...")

	return k.apply(c, ui, options, true)
}

func (k Carrier) createCarrierNamespace(c *kubernetes.Cluster, ui *termui.UI) error {
	_, err := c.Kubectl.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: CarrierDeploymentID,
				Labels: map[string]string{
					kubernetes.CarrierDeploymentLabelKey: kubernetes.CarrierDeploymentLabelValue,
				},
			},
		},
		metav1.CreateOptions{},
	)

	return err
}

// Replaces {{version}} with version.Version and applies the embedded yaml
func applyCarrierServerYaml(c *kubernetes.Cluster, ui *termui.UI) (string, error) {
	yamlPathOnDisk, err := helpers.ExtractFile(carrierServerYaml)
	if err != nil {
		return "", errors.New("Failed to extract embedded file: " + carrierServerYaml + " - " + err.Error())
	}
	defer os.Remove(yamlPathOnDisk)

	fileContents, err := ioutil.ReadFile(yamlPathOnDisk)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`{{version}}`)
	renderedFileContents := re.ReplaceAll(fileContents, []byte(version.Version))

	tmpFilePath, err := helpers.CreateTmpFile(string(renderedFileContents))
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFilePath)

	return helpers.Kubectl(fmt.Sprintf("apply -n %s --filename %s", CarrierDeploymentID, tmpFilePath))
}
