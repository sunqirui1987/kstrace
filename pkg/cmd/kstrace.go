package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"

	"github.com/suiqirui1987/kstrace/pkg/factory"
	"github.com/suiqirui1987/kstrace/pkg/strace"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

var (
	ktraceExamples = `
	kstrace -p xxxx -n default -f cpuwalk.bt
	`
)

type KStrace struct {
	configFlags *genericclioptions.ConfigFlags

	genericclioptions.IOStreams

	namespace string
	container string
	pod       string
	file      string
	code      string
	nodeName  string

	program string

	pod_v  *v1.Pod
	node_v *v1.Node

	clientConfig *rest.Config
}

func NewKStrace(streams genericclioptions.IOStreams) *KStrace {
	return &KStrace{
		configFlags: genericclioptions.NewConfigFlags(false),

		IOStreams: streams,
	}
}

func NewKStraceCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewKStrace(streams)

	cmd := &cobra.Command{
		Use:          "kstrace",
		SilenceUsage: true,
		Short:        `xxxxxx`,                               // Wrap with i18n.T()
		Example:      fmt.Sprintf(ktraceExamples, "kubectl"), // Wrap with templates.Examples()
		RunE: func(c *cobra.Command, args []string) error {

			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.file, "filename", "f", o.file, "bpftrace File")
	cmd.Flags().StringVarP(&o.code, "bpftracecode", "e", o.code, "bpftrace code")
	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", o.namespace, "Specify namespace")
	cmd.Flags().StringVarP(&o.pod, "pod", "p", o.pod, "Specify pod")
	cmd.Flags().StringVarP(&o.container, "container", "c", o.container, "Specify container")
	return cmd
}

func (o *KStrace) Validate() error {

	if o.namespace == "" {
		return errors.New("namespace value is empty should be custom or default")
	}

	return nil
}
func (o *KStrace) Complete(cmd *cobra.Command, args []string) error {

	log.Info("running in verbose mode")
	log.SetLevel(log.DebugLevel)

	flags := cmd.PersistentFlags()
	o.configFlags.AddFlags(flags)

	matchVersionFlags := factory.NewMatchVersionFlags(o.configFlags)
	matchVersionFlags.AddFlags(flags)

	f := factory.NewFactory(matchVersionFlags)

	// Prepare program
	if len(o.file) > 0 {
		b, err := ioutil.ReadFile(o.file)
		if err != nil {
			return fmt.Errorf("error opening program file")
		}
		o.program = string(b)
	} else {
		o.program = o.code
	}

	//get pod res
	pod_r, err := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.namespace).
		SingleResourceType().
		ResourceNames("pods", o.pod).
		Do().Object()
	if err != nil {
		return err
	}
	o.pod_v, _ = pod_r.(*v1.Pod)

	found := false
	for _, c := range o.pod_v.Spec.Containers {

		if o.container == "" {
			o.container = c.Name
			found = true
			break
		}

		if c.Name == o.container {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no containers found for the provided pod/container combination")
	}

	//get node res
	node_r, err := f.
		NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ResourceNames("nodes", o.pod_v.Spec.NodeName).
		Do().Object()

	if err != nil {
		return err
	}
	o.node_v, _ = node_r.(*v1.Node)

	// get lablels
	labels := o.node_v.GetLabels()
	val, ok := labels["kubernetes.io/hostname"]
	if !ok {
		return fmt.Errorf("label kubernetes.io/hostname not found in node")
	}
	o.nodeName = val

	// Prepare client
	o.clientConfig, err = f.ToRESTConfig()
	if err != nil {
		return err
	}

	return nil

}

func (o *KStrace) Run() error {
	juid := uuid.NewUUID()
	jobsClient, err := batchv1client.NewForConfig(o.clientConfig)
	if err != nil {
		return err
	}

	coreClient, err := corev1client.NewForConfig(o.clientConfig)
	if err != nil {
		return err
	}

	sc := &strace.StraceJobClient{
		JobClient:    jobsClient.Jobs(o.namespace),
		ConfigClient: coreClient.ConfigMaps(o.namespace),
	}

	sj := strace.StraceJob{
		Name:      fmt.Sprintf("%s%s", strace.TracePrefix, string(juid)),
		Namespace: o.namespace,
		ID:        juid,

		Hostname:      o.nodeName,
		Program:       o.program,
		PodUID:        string(o.pod_v.UID),
		PodName:       o.pod,
		ContainerName: o.container,
	}

	job, err := sc.CreateJob(sj)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.IOStreams.Out, "kstrace %s created\n", sj.ID)

	ctx := context.Background()
	ctx = strace.WithStandardSignals(ctx)
	a := strace.NewAttacher(coreClient, o.clientConfig, o.IOStreams)
	a.WithContext(ctx)
	a.AttachJob(sj.ID, job.Namespace)

	return nil
}
