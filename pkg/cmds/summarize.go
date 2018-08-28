package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/appscode/go/net/httpclient"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/cli/pkg/kube"
	"github.com/kubedb/cli/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdSummarize(out io.Writer, cmdErr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Export summary report",
		Run: func(cmd *cobra.Command, args []string) {
			f := kube.NewKubeFactory(cmd)
			cmdutil.CheckErr(exportReport(f, cmd, out, cmdErr, args))
		},
	}
	util.AddAuditReportFlags(cmd)
	return cmd
}

const (
	validResourcesForReport = `Valid resource types include:

    * elasticsearches
    * postgreses
    * mysqls
    * mongodbs
    * redises
    * memcacheds
    * etcds
    `
)

var operatorLabel = map[string]string{
	"app": "kubedb",
}

const (
	operatorPortNumber = 8080
)

func exportReport(f cmdutil.Factory, cmd *cobra.Command, out, errOut io.Writer, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(errOut, "You must specify the type of resource. ", validResourcesForReport)
		usageString := "Required resource not specified."
		return cmdutil.UsageErrorf(cmd, usageString)
	}

	if len(strings.Split(args[0], ",")) > 1 {
		return errors.New("audit doesn't support multiple resource")
	}

	resource := args[0]
	items := strings.Split(resource, "/")
	kubedbType, err := util.GetResourceType(items[0])
	if err != nil {
		return err
	}

	switch kubedbType {
	case tapi.ResourcePluralSnapshot, tapi.ResourcePluralDormantDatabase:
		return fmt.Errorf(`Failed to summarize resource type "%v"`, items[0])
	}

	var kubedbName string
	if len(items) > 1 {
		if len(items) > 2 {
			return errors.New("Only one database can be summarized at a time.")
		}
		kubedbName = items[1]
	} else {
		if len(args) > 2 {
			return errors.New("Only one database can be summarized at a time.")
		}
		kubedbName = args[1]
	}

	namespace, _ := util.GetNamespace(cmd)

	goClient, err := f.ClientSet()
	if err != nil {
		return err
	}

	operatorNamespace := cmdutil.GetFlagString(cmd, "operator-namespace")
	operatorPodList, err := goClient.Core().Pods(operatorNamespace).List(
		metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(
				operatorLabel,
			).String(),
		},
	)
	if err != nil {
		return err
	}

	if len(operatorPodList.Items) == 0 {
		return errors.New("Operator pod not found")
	}

	restClient, err := f.RESTClient()
	if err != nil {
		return err
	}

	config, err := f.ClientConfig()
	if err != nil {
		return err
	}

	tunnel := newTunnel(restClient, config, operatorNamespace, operatorPodList.Items[0].Name, operatorPortNumber)
	if err := tunnel.forwardPort(); err != nil {
		return err
	}

	proxyClient := httpclient.Default().WithBaseURL(fmt.Sprintf("http://127.0.0.1:%d", tunnel.Local))
	summaryReportURL, _ := url.Parse(fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%v/%v/%v/report", namespace, kubedbType, kubedbName))

	index := cmdutil.GetFlagString(cmd, "index")
	if index != "" {
		summaryReportURL.Query().Set("index", index)
	}
	req, err := proxyClient.NewRequest("GET", summaryReportURL.String(), nil)
	if err != nil {
		return err
	}

	var report *tapi.Report
	if _, err := proxyClient.Do(req, &report); err != nil {
		return err
	}

	outputDirectory := cmdutil.GetFlagString(cmd, "output")
	fileName := fmt.Sprintf("report-%v.json", time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join(outputDirectory, fileName)

	reportDataByte, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := util.WriteJson(path, reportDataByte); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf(`Summary report for "%v/%v" has been stored in '%v'`, kubedbType, kubedbName, path))
	return nil
}

type tunnel struct {
	Local     int
	Remote    int
	Namespace string
	PodName   string
	Out       io.Writer
	stopChan  chan struct{}
	readyChan chan struct{}
	config    *rest.Config
	client    rest.Interface
}

func newTunnel(client rest.Interface, config *rest.Config, namespace, podName string, remote int) *tunnel {
	return &tunnel{
		config:    config,
		client:    client,
		Namespace: namespace,
		PodName:   podName,
		Remote:    remote,
		stopChan:  make(chan struct{}, 1),
		readyChan: make(chan struct{}, 1),
		Out:       ioutil.Discard,
	}
}

func (t *tunnel) forwardPort() error {
	u := t.client.Post().
		Resource("pods").
		Namespace(t.Namespace).
		Name(t.PodName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(t.config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)

	local, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("could not find an available port: %s", err)
	}
	t.Local = local

	ports := []string{fmt.Sprintf("%d:%d", t.Local, t.Remote)}

	pf, err := portforward.New(dialer, ports, t.stopChan, t.readyChan, t.Out, t.Out)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return fmt.Errorf("forwarding ports: %v", err)
	case <-pf.Ready:
		return nil
	}
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}
